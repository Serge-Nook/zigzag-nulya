"""Обработчик OLE / старых форматов (doc, xls)."""

from __future__ import annotations

from pathlib import Path

import olefile

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata

_AUTHOR_PROP = 0x04  # PIDSI_AUTHOR
_LAST_SAVED_BY_PROP = 0x08  # PIDSI_LASTSAVEDBY


class OLEHandler(BaseHandler):
    """Обработчик бинарных OLE-форматов (doc, xls)."""

    SUPPORTED_EXTENSIONS = (".doc", ".xls")

    def get_format_group(self) -> str:
        return "OLE / Старые форматы"

    def read_metadata(self, path: Path) -> FileMetadata:
        meta = self._build_base_metadata(path)
        author, last_mod = self.read_author_fields(path)
        meta.author = author
        meta.last_modified_by = last_mod
        return meta

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        try:
            ole = olefile.OleFileIO(str(path))
        except Exception:
            return ("", "")

        author = ""
        last_mod = ""
        try:
            meta = ole.get_metadata()
            if meta.author:
                author = (
                    meta.author.decode("utf-8", errors="replace")
                    if isinstance(meta.author, bytes)
                    else str(meta.author)
                )
            if meta.last_saved_by:
                last_mod = (
                    meta.last_saved_by.decode("utf-8", errors="replace")
                    if isinstance(meta.last_saved_by, bytes)
                    else str(meta.last_saved_by)
                )
        except Exception:
            pass
        finally:
            ole.close()

        return (author, last_mod)

    def write_author(
        self,
        path: Path,
        author: str | None,
        last_modified_by: str | None,
    ) -> None:
        if author is None and last_modified_by is None:
            return

        try:
            ole = olefile.OleFileIO(str(path), write_mode=True)
        except Exception as exc:
            raise OSError(f"Не удалось открыть OLE-файл: {exc}") from exc

        try:
            meta = ole.get_metadata()

            if author is not None:
                meta.author = author.encode("utf-8") if author else None

            if last_modified_by is not None:
                meta.last_saved_by = (
                    last_modified_by.encode("utf-8") if last_modified_by else None
                )

            ole.write_metadata(meta)
        except Exception as exc:
            raise OSError(f"Ошибка записи метаданных OLE: {exc}") from exc
        finally:
            ole.close()
