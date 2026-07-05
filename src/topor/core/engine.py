"""Движок обработки файлов — оркестратор."""

from __future__ import annotations

import logging
from pathlib import Path

from PyQt6.QtCore import QObject, QThread, pyqtSignal

from topor.core.handlers import HANDLER_MAP, BaseHandler
from topor.core.models import (
    AuthorAction,
    DateMode,
    FileMetadata,
    ProcessingResult,
    ProcessingSettings,
    ProcessingStatus,
)
from topor.core.timestamps import apply_timestamps
from topor.utils.backup import create_backup

logger = logging.getLogger("topor.engine")

SIZE_WARNING_BYTES = 500 * 1024 * 1024  # 500 МБ


def get_handler(path: Path) -> BaseHandler | None:
    """Получить обработчик для файла по расширению."""
    ext = path.suffix.lower()
    handler_cls = HANDLER_MAP.get(ext)
    if handler_cls is None:
        return None
    return handler_cls()


def read_file_metadata(path: Path) -> FileMetadata | None:
    """Прочитать метаданные одного файла."""
    handler = get_handler(path)
    if handler is None:
        return None
    try:
        return handler.read_metadata(path)
    except Exception as exc:
        logger.warning("Ошибка чтения метаданных %s: %s", path, exc)
        return None


class ProcessingWorker(QObject):
    """Рабочий поток обработки файлов."""

    progress = pyqtSignal(int, int)  # (текущий, всего)
    file_done = pyqtSignal(object)  # ProcessingResult
    finished = pyqtSignal()
    log_message = pyqtSignal(str)

    def __init__(
        self,
        files: list[Path],
        settings: ProcessingSettings,
        parent: QObject | None = None,
    ) -> None:
        super().__init__(parent)
        self._files = files
        self._settings = settings
        self._cancelled = False

    def cancel(self) -> None:
        self._cancelled = True

    def run(self) -> None:
        total = len(self._files)
        for i, path in enumerate(self._files):
            if self._cancelled:
                self.log_message.emit("Обработка отменена пользователем.")
                break

            result = self._process_file(path)
            self.file_done.emit(result)
            self.progress.emit(i + 1, total)

        self.finished.emit()

    def _process_file(self, path: Path) -> ProcessingResult:
        result = ProcessingResult(path=path)

        if not path.exists():
            result.status = ProcessingStatus.ERROR
            result.message = "Файл не найден"
            self.log_message.emit(f"ОШИБКА: {path} — файл не найден")
            return result

        if not path.is_file():
            result.status = ProcessingStatus.SKIPPED
            result.message = "Не является файлом"
            return result

        handler = get_handler(path)
        if handler is None:
            result.status = ProcessingStatus.SKIPPED
            result.message = "Формат не поддерживается"
            self.log_message.emit(f"ПРОПУСК: {path} — формат не поддерживается")
            return result

        try:
            if not _check_access(path):
                result.status = ProcessingStatus.SKIPPED
                result.message = "Нет доступа к файлу"
                self.log_message.emit(f"ПРОПУСК: {path} — нет доступа")
                return result

            if path.stat().st_size > SIZE_WARNING_BYTES:
                self.log_message.emit(
                    f"ВНИМАНИЕ: {path} — размер > 500 МБ"
                )

            if self._settings.create_backup:
                bak = create_backup(path)
                self.log_message.emit(f"Резервная копия: {bak}")

            self._apply_timestamps(path)
            self._apply_author(path, handler)

            result.status = ProcessingStatus.SUCCESS
            result.message = "Обработан успешно"
            self.log_message.emit(f"ОК: {path}")

        except PermissionError:
            result.status = ProcessingStatus.SKIPPED
            result.message = "Файл заблокирован"
            self.log_message.emit(f"ПРОПУСК: {path} — файл заблокирован")
        except Exception as exc:
            result.status = ProcessingStatus.ERROR
            result.message = str(exc)
            self.log_message.emit(f"ОШИБКА: {path} — {exc}")

        return result

    def _apply_timestamps(self, path: Path) -> None:
        ts = self._settings.time_settings
        if ts.creation_mode == DateMode.NONE and ts.modification_mode == DateMode.NONE:
            return
        apply_timestamps(path, ts)

    def _apply_author(self, path: Path, handler: BaseHandler) -> None:
        auth = self._settings.author_settings
        if (
            auth.author_action == AuthorAction.NONE
            and auth.last_modified_by_action == AuthorAction.NONE
        ):
            return

        author_val: str | None = None
        lmb_val: str | None = None

        if auth.author_action == AuthorAction.DELETE:
            author_val = ""
        elif auth.author_action == AuthorAction.OVERWRITE:
            author_val = auth.new_author

        if auth.last_modified_by_action == AuthorAction.DELETE:
            lmb_val = ""
        elif auth.last_modified_by_action == AuthorAction.OVERWRITE:
            lmb_val = auth.new_last_modified_by

        handler.write_author(path, author_val, lmb_val)


def _check_access(path: Path) -> bool:
    """Проверить доступ на чтение и запись."""
    import os

    return os.access(str(path), os.R_OK | os.W_OK)


class EngineThread(QThread):
    """Поток-обёртка для ProcessingWorker."""

    def __init__(
        self,
        worker: ProcessingWorker,
        parent: QObject | None = None,
    ) -> None:
        super().__init__(parent)
        self._worker = worker

    def run(self) -> None:
        self._worker.run()
