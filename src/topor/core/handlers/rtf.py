"""Обработчик RTF-файлов."""

from __future__ import annotations

import re
from pathlib import Path

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata

AUTHOR_RE = re.compile(r"\{\\author\s+(.*?)\}", re.DOTALL)
OPERATOR_RE = re.compile(r"\{\\operator\s+(.*?)\}", re.DOTALL)


class RTFHandler(BaseHandler):
    """Обработчик RTF-файлов."""

    SUPPORTED_EXTENSIONS = (".rtf",)

    def get_format_group(self) -> str:
        return "RTF"

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
        match = AUTHOR_RE.search(content)
        if match:
            author = match.group(1).strip()

        last_mod = ""
        match = OPERATOR_RE.search(content)
        if match:
            last_mod = match.group(1).strip()

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
            content = self._update_field(content, AUTHOR_RE, "author", author)

        if last_modified_by is not None:
            content = self._update_field(
                content, OPERATOR_RE, "operator", last_modified_by
            )

        path.write_text(content, encoding="utf-8")

    def _update_field(
        self,
        content: str,
        pattern: re.Pattern[str],
        field_name: str,
        value: str,
    ) -> str:
        if value == "":
            return pattern.sub("", content)

        replacement = "{\\%s %s}" % (field_name, value)

        if pattern.search(content):
            return pattern.sub(replacement.replace("\\", "\\\\"), content)

        info_match = re.search(r"\{\\info\b", content)
        if info_match:
            insert_pos = info_match.end()
            return content[:insert_pos] + replacement + content[insert_pos:]

        return content
