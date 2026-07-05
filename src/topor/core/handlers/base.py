"""Базовый обработчик метаданных файлов."""

from __future__ import annotations

import abc
from datetime import datetime
from pathlib import Path

from topor.core.models import FileMetadata


class BaseHandler(abc.ABC):
    """Абстрактный базовый класс обработчика формата."""

    SUPPORTED_EXTENSIONS: tuple[str, ...] = ()

    def supports(self, path: Path) -> bool:
        return path.suffix.lower() in self.SUPPORTED_EXTENSIONS

    @abc.abstractmethod
    def read_metadata(self, path: Path) -> FileMetadata:
        """Считать текущие метаданные файла."""

    @abc.abstractmethod
    def write_author(
        self,
        path: Path,
        author: str | None,
        last_modified_by: str | None,
    ) -> None:
        """Записать / удалить поля автора.

        None означает «не менять», пустая строка — «удалить».
        """

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        """Вернуть (author, last_modified_by). По умолчанию пустые."""
        return ("", "")

    def can_modify_author(self) -> bool:
        """Поддерживает ли формат изменение автора внутри файла."""
        return True

    def get_format_group(self) -> str:
        """Название группы форматов."""
        return "Неизвестный"

    def _read_fs_times(self, path: Path) -> tuple[datetime | None, datetime | None]:
        """Прочитать временные штампы файловой системы."""
        try:
            stat = path.stat()
            ctime = datetime.fromtimestamp(stat.st_ctime)
            mtime = datetime.fromtimestamp(stat.st_mtime)
            return ctime, mtime
        except OSError:
            return None, None

    def _build_base_metadata(self, path: Path) -> FileMetadata:
        """Создать базовый объект метаданных с данными ФС."""
        ctime, mtime = self._read_fs_times(path)
        try:
            size = path.stat().st_size
        except OSError:
            size = 0
        return FileMetadata(
            path=path,
            size=size,
            creation_time=ctime,
            modification_time=mtime,
            format_group=self.get_format_group(),
        )
