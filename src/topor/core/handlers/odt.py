"""Обработчик формата ODT (OpenDocument Text)."""

from __future__ import annotations

import zipfile
from io import BytesIO
from pathlib import Path

from lxml import etree

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata

META_XML_PATH = "meta.xml"
OFFICE_NS = "urn:oasis:names:tc:opendocument:xmlns:office:1.0"
META_NS = "urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
DC_NS = "http://purl.org/dc/elements/1.1/"


class ODTHandler(BaseHandler):
    """Обработчик формата OpenDocument."""

    SUPPORTED_EXTENSIONS = (".odt",)

    def get_format_group(self) -> str:
        return "OpenDocument"

    def read_metadata(self, path: Path) -> FileMetadata:
        meta = self._build_base_metadata(path)
        author, last_mod = self.read_author_fields(path)
        meta.author = author
        meta.last_modified_by = last_mod
        return meta

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        try:
            root = self._read_meta_xml(path)
        except (zipfile.BadZipFile, KeyError, OSError):
            return ("", "")

        if root is None:
            return ("", "")

        office_meta = root.find(f"{{{OFFICE_NS}}}meta")
        if office_meta is None:
            return ("", "")

        author = self._get_text(office_meta, f"{{{META_NS}}}initial-creator")
        if not author:
            author = self._get_text(office_meta, f"{{{DC_NS}}}creator")

        last_mod = self._get_text(office_meta, f"{{{DC_NS}}}creator")
        return (author, last_mod)

    def write_author(
        self,
        path: Path,
        author: str | None,
        last_modified_by: str | None,
    ) -> None:
        if author is None and last_modified_by is None:
            return

        buf = BytesIO(path.read_bytes())
        with zipfile.ZipFile(buf, "r") as zr:
            names = zr.namelist()
            if META_XML_PATH not in names:
                return
            meta_bytes = zr.read(META_XML_PATH)

        root = etree.fromstring(meta_bytes)
        office_meta = root.find(f"{{{OFFICE_NS}}}meta")
        if office_meta is None:
            return

        if author is not None:
            self._set_or_remove(
                office_meta, f"{{{META_NS}}}initial-creator", author
            )

        if last_modified_by is not None:
            self._set_or_remove(
                office_meta, f"{{{DC_NS}}}creator", last_modified_by
            )

        new_meta = etree.tostring(
            root, xml_declaration=True, encoding="UTF-8"
        )

        out = BytesIO()
        with zipfile.ZipFile(buf, "r") as zr:
            with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as zw:
                for item in zr.infolist():
                    if item.filename == META_XML_PATH:
                        zw.writestr(item, new_meta)
                    else:
                        zw.writestr(item, zr.read(item.filename))

        path.write_bytes(out.getvalue())

    def _read_meta_xml(self, path: Path) -> etree._Element | None:
        with zipfile.ZipFile(str(path), "r") as zf:
            if META_XML_PATH not in zf.namelist():
                return None
            data = zf.read(META_XML_PATH)
            return etree.fromstring(data)

    def _get_text(self, parent: etree._Element, tag: str) -> str:
        elem = parent.find(tag)
        if elem is not None and elem.text:
            return elem.text.strip()
        return ""

    def _set_or_remove(
        self, parent: etree._Element, tag: str, value: str
    ) -> None:
        elem = parent.find(tag)
        if value == "":
            if elem is not None:
                parent.remove(elem)
        else:
            if elem is None:
                elem = etree.SubElement(parent, tag)
            elem.text = value
