"""Обработчик HTML-файлов."""

from __future__ import annotations

import re
from pathlib import Path

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata

AUTHOR_META_RE = re.compile(
    r'(<meta\s+name\s*=\s*"author"\s+content\s*=\s*")([^"]*?)("\s*/?>)',
    re.IGNORECASE,
)

LAST_MOD_META_RE = re.compile(
    r'(<meta\s+name\s*=\s*"last-modified-by"\s+content\s*=\s*")([^"]*?)("\s*/?>)',
    re.IGNORECASE,
)


class HTMLHandler(BaseHandler):
    """Обработчик HTML-файлов."""

    SUPPORTED_EXTENSIONS = (".html", ".htm")

    def get_format_group(self) -> str:
        return "HTML"

    def read_metadata(self, path: Path) -> FileMetadata:
        meta = self._build_base_metadata(path)
        author, last_mod = self.read_author_fields(path)
        meta.author = author
        meta.last_modified_by = last_mod
        return meta

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        try:
            content = path.read_text(encoding="utf-8", errors="replace")
        except OSError:
            return ("", "")

        author = ""
        match = AUTHOR_META_RE.search(content)
        if match:
            author = match.group(2)

        last_mod = ""
        match = LAST_MOD_META_RE.search(content)
        if match:
            last_mod = match.group(2)

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
            content = path.read_text(encoding="utf-8", errors="replace")
        except OSError:
            return

        if author is not None:
            content = self._update_meta(
                content, AUTHOR_META_RE, "author", author
            )

        if last_modified_by is not None:
            content = self._update_meta(
                content, LAST_MOD_META_RE, "last-modified-by", last_modified_by
            )

        path.write_text(content, encoding="utf-8")

    def _update_meta(
        self, content: str, pattern: re.Pattern[str], name: str, value: str
    ) -> str:
        if value == "":
            return pattern.sub("", content)

        if pattern.search(content):
            return pattern.sub(rf"\g<1>{re.escape(value)}\g<3>", content)

        tag = f'<meta name="{name}" content="{value}" />'
        head_match = re.search(r"(<head[^>]*>)", content, re.IGNORECASE)
        if head_match:
            insert_pos = head_match.end()
            return content[:insert_pos] + "\n" + tag + content[insert_pos:]

        return tag + "\n" + content
