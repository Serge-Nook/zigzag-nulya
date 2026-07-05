"""Обработчик текстовых файлов (txt, xml)."""

from __future__ import annotations

from pathlib import Path

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata


class PlainTextHandler(BaseHandler):
    """Обработчик простых текстовых файлов.

    Для txt и xml изменение автора возможно только через
    расширенные атрибуты файловой системы (xattr).
    """

    SUPPORTED_EXTENSIONS = (".txt", ".xml")

    def get_format_group(self) -> str:
        return "Текст / Разметка"

    def can_modify_author(self) -> bool:
        return False

    def read_metadata(self, path: Path) -> FileMetadata:
        meta = self._build_base_metadata(path)
        author, last_mod = self._read_xattr(path)
        meta.author = author
        meta.last_modified_by = last_mod
        return meta

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        return self._read_xattr(path)

    def write_author(
        self,
        path: Path,
        author: str | None,
        last_modified_by: str | None,
    ) -> None:
        if author is not None:
            self._write_xattr(path, "user.author", author)
        if last_modified_by is not None:
            self._write_xattr(path, "user.last_modified_by", last_modified_by)

    def _read_xattr(self, path: Path) -> tuple[str, str]:
        """Прочитать расширенные атрибуты ФС."""
        author = self._get_xattr(path, "user.author")
        last_mod = self._get_xattr(path, "user.last_modified_by")
        return (author, last_mod)

    def _get_xattr(self, path: Path, attr_name: str) -> str:
        try:
            import xattr

            val = xattr.getxattr(str(path), attr_name)
            return val.decode("utf-8", errors="replace")
        except Exception:
            try:
                import os

                val = os.getxattr(str(path), attr_name)
                return val.decode("utf-8", errors="replace")
            except Exception:
                return ""

    def _write_xattr(self, path: Path, attr_name: str, value: str) -> None:
        try:
            import os

            if value:
                os.setxattr(str(path), attr_name, value.encode("utf-8"))
            else:
                try:
                    os.removexattr(str(path), attr_name)
                except OSError:
                    pass
        except Exception:
            pass
