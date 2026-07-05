"""Таблица предпросмотра файлов."""

from __future__ import annotations

from pathlib import Path

from PyQt6.QtCore import Qt
from PyQt6.QtWidgets import (
    QAbstractItemView,
    QHeaderView,
    QTableWidget,
    QTableWidgetItem,
    QWidget,
)

from topor.core.models import FileMetadata

COLUMNS = [
    "Файл",
    "Путь",
    "Размер",
    "Формат",
    "Дата создания",
    "Дата изменения",
    "Автор",
    "Последний редактор",
]


class FileTable(QTableWidget):
    """Таблица предпросмотра метаданных файлов."""

    def __init__(self, parent: QWidget | None = None) -> None:
        super().__init__(parent)
        self.setColumnCount(len(COLUMNS))
        self.setHorizontalHeaderLabels(COLUMNS)
        self.setAlternatingRowColors(True)
        self.setSelectionBehavior(QAbstractItemView.SelectionBehavior.SelectRows)
        self.setEditTriggers(QAbstractItemView.EditTrigger.NoEditTriggers)
        self.setSortingEnabled(True)
        self.verticalHeader().setDefaultSectionSize(26)

        header = self.horizontalHeader()
        header.setSectionResizeMode(0, QHeaderView.ResizeMode.Interactive)
        header.setSectionResizeMode(1, QHeaderView.ResizeMode.Stretch)
        header.setStretchLastSection(True)
        self.setColumnWidth(0, 200)
        self.setColumnWidth(2, 90)
        self.setColumnWidth(3, 130)
        self.setColumnWidth(4, 150)
        self.setColumnWidth(5, 150)
        self.setColumnWidth(6, 130)
        self.setColumnWidth(7, 130)

        self._metadata: list[FileMetadata] = []

    def load_metadata(self, metadata_list: list[FileMetadata]) -> None:
        """Заполнить таблицу метаданными."""
        self._metadata = metadata_list
        self.setSortingEnabled(False)
        self.setRowCount(len(metadata_list))

        for row, meta in enumerate(metadata_list):
            self._set_row(row, meta)

        self.setSortingEnabled(True)

    def _set_row(self, row: int, meta: FileMetadata) -> None:
        self.setItem(row, 0, QTableWidgetItem(meta.filename))
        self.setItem(row, 1, QTableWidgetItem(str(meta.path.parent)))
        self.setItem(row, 2, QTableWidgetItem(meta.size_display))
        self.setItem(row, 3, QTableWidgetItem(meta.format_group))
        self.setItem(
            row,
            4,
            QTableWidgetItem(
                meta.creation_time.strftime("%d.%m.%Y %H:%M:%S")
                if meta.creation_time
                else ""
            ),
        )
        self.setItem(
            row,
            5,
            QTableWidgetItem(
                meta.modification_time.strftime("%d.%m.%Y %H:%M:%S")
                if meta.modification_time
                else ""
            ),
        )
        self.setItem(row, 6, QTableWidgetItem(meta.author))
        self.setItem(row, 7, QTableWidgetItem(meta.last_modified_by))

    def get_files(self) -> list[Path]:
        """Вернуть список путей из загруженных метаданных."""
        return [m.path for m in self._metadata]

    def clear_all(self) -> None:
        """Очистить таблицу."""
        self._metadata.clear()
        self.setRowCount(0)

    def update_status(self, path: Path, status: str) -> None:
        """Обновить статус файла в таблице (подсветка строки)."""
        for row, meta in enumerate(self._metadata):
            if meta.path == path:
                color_map = {
                    "success": Qt.GlobalColor.green,
                    "error": Qt.GlobalColor.red,
                    "skipped": Qt.GlobalColor.yellow,
                }
                color = color_map.get(status)
                if color:
                    for col in range(self.columnCount()):
                        item = self.item(row, col)
                        if item:
                            item.setBackground(color)
                break
