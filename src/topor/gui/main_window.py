"""Главное окно приложения Топор."""

from __future__ import annotations

import logging
from pathlib import Path

from PyQt6.QtCore import QMimeData, Qt, QUrl
from PyQt6.QtWidgets import (
    QCheckBox,
    QComboBox,
    QFileDialog,
    QGroupBox,
    QHBoxLayout,
    QLabel,
    QLineEdit,
    QMainWindow,
    QMessageBox,
    QProgressBar,
    QPushButton,
    QSplitter,
    QStatusBar,
    QTabWidget,
    QVBoxLayout,
    QWidget,
)

from topor import __version__, __website__
from topor.core.engine import (
    EngineThread,
    ProcessingWorker,
    read_file_metadata,
)
from topor.core.file_discovery import SUPPORTED_EXTENSIONS, discover_files
from topor.core.models import (
    FileMetadata,
    ProcessingResult,
    ProcessingSettings,
    ProcessingStatus,
)
from topor.gui.author_panel import AuthorPanel
from topor.gui.file_table import FileTable
from topor.gui.log_panel import LogPanel
from topor.gui.time_panel import TimePanel

logger = logging.getLogger("topor.gui")


class MainWindow(QMainWindow):
    """Главное окно программы Топор."""

    def __init__(self) -> None:
        super().__init__()
        self.setWindowTitle(f"Топор — Массовый редактор метаданных v{__version__}")
        self.setMinimumSize(1000, 700)
        self.resize(1200, 800)
        self.setAcceptDrops(True)

        self._worker: ProcessingWorker | None = None
        self._thread: EngineThread | None = None
        self._files: list[Path] = []

        self._build_ui()
        self._connect_signals()
        self._update_status_bar()

    # ─── UI Construction ──────────────────────────────────────

    def _build_ui(self) -> None:
        central = QWidget()
        self.setCentralWidget(central)
        main_layout = QVBoxLayout(central)

        # Верхняя панель: выбор файлов
        file_group = QGroupBox("Входные данные")
        file_layout = QVBoxLayout(file_group)

        btn_row = QHBoxLayout()
        self._btn_add_files = QPushButton("Добавить файлы...")
        self._btn_add_folder = QPushButton("Добавить папку...")
        self._btn_clear = QPushButton("Очистить список")
        self._btn_clear.setObjectName("secondaryButton")

        btn_row.addWidget(self._btn_add_files)
        btn_row.addWidget(self._btn_add_folder)
        btn_row.addStretch()
        btn_row.addWidget(self._btn_clear)
        file_layout.addLayout(btn_row)

        options_row = QHBoxLayout()
        self._chk_recursive = QCheckBox("Включая подкаталоги")
        options_row.addWidget(self._chk_recursive)

        options_row.addWidget(QLabel("Фильтр по маске:"))
        self._mask_input = QLineEdit()
        self._mask_input.setPlaceholderText("Например: *.docx;*.xlsx")
        self._mask_input.setMaximumWidth(300)
        options_row.addWidget(self._mask_input)

        self._chk_backup = QCheckBox("Создавать .bak копии")
        options_row.addWidget(self._chk_backup)
        options_row.addStretch()
        file_layout.addLayout(options_row)

        main_layout.addWidget(file_group)

        # Разделитель: таблица + настройки
        splitter = QSplitter(Qt.Orientation.Vertical)

        # Таблица предпросмотра
        self._file_table = FileTable()
        splitter.addWidget(self._file_table)

        # Вкладки настроек + лог
        bottom_widget = QWidget()
        bottom_layout = QVBoxLayout(bottom_widget)
        bottom_layout.setContentsMargins(0, 0, 0, 0)

        self._tabs = QTabWidget()
        self._time_panel = TimePanel()
        self._author_panel = AuthorPanel()
        self._log_panel = LogPanel()
        self._tabs.addTab(self._time_panel, "Временные штампы")
        self._tabs.addTab(self._author_panel, "Метаданные автора")
        self._tabs.addTab(self._log_panel, "Журнал операций")
        bottom_layout.addWidget(self._tabs)

        # Прогресс-бар и кнопки действий
        action_row = QHBoxLayout()
        self._progress_bar = QProgressBar()
        self._progress_bar.setTextVisible(True)
        self._progress_bar.setValue(0)
        action_row.addWidget(self._progress_bar, stretch=1)

        self._btn_apply = QPushButton("Применить")
        self._btn_apply.setFixedWidth(140)
        action_row.addWidget(self._btn_apply)

        self._btn_cancel = QPushButton("Отмена")
        self._btn_cancel.setObjectName("dangerButton")
        self._btn_cancel.setFixedWidth(100)
        self._btn_cancel.setEnabled(False)
        action_row.addWidget(self._btn_cancel)

        bottom_layout.addLayout(action_row)
        splitter.addWidget(bottom_widget)

        splitter.setSizes([350, 350])
        main_layout.addWidget(splitter)

        # Статус-бар
        self._status_bar = QStatusBar()
        self.setStatusBar(self._status_bar)

    # ─── Signals ──────────────────────────────────────────────

    def _connect_signals(self) -> None:
        self._btn_add_files.clicked.connect(self._on_add_files)
        self._btn_add_folder.clicked.connect(self._on_add_folder)
        self._btn_clear.clicked.connect(self._on_clear)
        self._btn_apply.clicked.connect(self._on_apply)
        self._btn_cancel.clicked.connect(self._on_cancel)

    # ─── File Selection ──────────────────────────────────────

    def _on_add_files(self) -> None:
        exts = " ".join(f"*{e}" for e in sorted(SUPPORTED_EXTENSIONS))
        paths, _ = QFileDialog.getOpenFileNames(
            self,
            "Выбор файлов",
            "",
            f"Поддерживаемые форматы ({exts});;Все файлы (*)",
        )
        if paths:
            self._add_paths([Path(p) for p in paths])

    def _on_add_folder(self) -> None:
        folder = QFileDialog.getExistingDirectory(self, "Выбор папки")
        if folder:
            self._add_paths([Path(folder)])

    def _add_paths(self, paths: list[Path]) -> None:
        mask = self._mask_input.text().strip()
        format_filter: set[str] = set()
        if mask:
            for part in mask.split(";"):
                part = part.strip().lstrip("*")
                if part:
                    format_filter.add(part if part.startswith(".") else f".{part}")

        new_files = discover_files(
            paths,
            include_subdirectories=self._chk_recursive.isChecked(),
            format_filter=format_filter or None,
        )
        existing = set(self._files)
        added = [f for f in new_files if f not in existing]
        self._files.extend(added)

        self._log_panel.append(
            f"Добавлено файлов: {len(added)} (всего: {len(self._files)})"
        )
        self._refresh_table()

    def _on_clear(self) -> None:
        self._files.clear()
        self._file_table.clear_all()
        self._progress_bar.setValue(0)
        self._update_status_bar()
        self._log_panel.append("Список файлов очищен.")

    def _refresh_table(self) -> None:
        metadata_list: list[FileMetadata] = []
        for path in self._files:
            meta = read_file_metadata(path)
            if meta:
                metadata_list.append(meta)
        self._file_table.load_metadata(metadata_list)
        self._update_status_bar()

    # ─── Drag & Drop ─────────────────────────────────────────

    def dragEnterEvent(self, event) -> None:  # type: ignore[override]
        if event.mimeData() and event.mimeData().hasUrls():
            event.acceptProposedAction()

    def dragMoveEvent(self, event) -> None:  # type: ignore[override]
        if event.mimeData() and event.mimeData().hasUrls():
            event.acceptProposedAction()

    def dropEvent(self, event) -> None:  # type: ignore[override]
        if event.mimeData() and event.mimeData().hasUrls():
            paths = [
                Path(url.toLocalFile())
                for url in event.mimeData().urls()
                if url.isLocalFile()
            ]
            if paths:
                self._add_paths(paths)
                event.acceptProposedAction()

    # ─── Processing ──────────────────────────────────────────

    def _on_apply(self) -> None:
        if not self._files:
            QMessageBox.warning(
                self,
                "Нет файлов",
                "Добавьте файлы для обработки.",
            )
            return

        settings = ProcessingSettings(
            time_settings=self._time_panel.get_settings(),
            author_settings=self._author_panel.get_settings(),
            create_backup=self._chk_backup.isChecked(),
            include_subdirectories=self._chk_recursive.isChecked(),
        )

        self._log_panel.append(
            f"Начало обработки: {len(self._files)} файлов..."
        )

        self._btn_apply.setEnabled(False)
        self._btn_cancel.setEnabled(True)
        self._progress_bar.setValue(0)
        self._progress_bar.setMaximum(len(self._files))

        self._worker = ProcessingWorker(list(self._files), settings)
        self._thread = EngineThread(self._worker)

        self._worker.progress.connect(self._on_progress)
        self._worker.file_done.connect(self._on_file_done)
        self._worker.finished.connect(self._on_finished)
        self._worker.log_message.connect(self._log_panel.append)

        self._thread.start()

    def _on_cancel(self) -> None:
        if self._worker:
            self._worker.cancel()
            self._log_panel.append("Отмена обработки...")

    def _on_progress(self, current: int, total: int) -> None:
        self._progress_bar.setValue(current)
        self._status_bar.showMessage(
            f"Обработано: {current} / {total}"
        )

    def _on_file_done(self, result: ProcessingResult) -> None:
        status_map = {
            ProcessingStatus.SUCCESS: "success",
            ProcessingStatus.ERROR: "error",
            ProcessingStatus.SKIPPED: "skipped",
        }
        status = status_map.get(result.status, "")
        self._file_table.update_status(result.path, status)

    def _on_finished(self) -> None:
        self._btn_apply.setEnabled(True)
        self._btn_cancel.setEnabled(False)
        self._log_panel.append("Обработка завершена.")
        self._tabs.setCurrentWidget(self._log_panel)
        self._refresh_table()
        self._update_status_bar()

        if self._thread:
            self._thread.quit()
            self._thread.wait()
            self._thread = None
            self._worker = None

    # ─── Status Bar ──────────────────────────────────────────

    def _update_status_bar(self) -> None:
        self._status_bar.showMessage(
            f"Файлов: {len(self._files)}  |  "
            f"Автор: Горшков С.В.  |  {__website__}"
        )
