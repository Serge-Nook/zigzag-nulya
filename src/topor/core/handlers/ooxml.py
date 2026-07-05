"""Обработчик Office Open XML (docx, xlsx, docm, xlsm)."""

from __future__ import annotations

import zipfile
from io import BytesIO
from pathlib import Path

from lxml import etree

from topor.core.handlers.base import BaseHandler
from topor.core.models import FileMetadata

CP_NS = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
DC_NS = "http://purl.org/dc/elements/1.1/"
DCTERMS_NS = "http://purl.org/dc/terms/"
XSI_NS = "http://www.w3.org/2001/XMLSchema-instance"

NSMAP = {
    "cp": CP_NS,
    "dc": DC_NS,
    "dcterms": DCTERMS_NS,
    "xsi": XSI_NS,
}

CORE_XML_PATH = "docProps/core.xml"


class OOXMLHandler(BaseHandler):
    """Обработчик форматов Office Open XML."""

    SUPPORTED_EXTENSIONS = (".docx", ".xlsx", ".docm", ".xlsm")

    def get_format_group(self) -> str:
        return "Office Open XML"

    def read_metadata(self, path: Path) -> FileMetadata:
        meta = self._build_base_metadata(path)
        author, last_mod = self.read_author_fields(path)
        meta.author = author
        meta.last_modified_by = last_mod
        return meta

    def read_author_fields(self, path: Path) -> tuple[str, str]:
        try:
            tree = self._read_core_xml(path)
        except (zipfile.BadZipFile, KeyError, OSError):
            return ("", "")

        if tree is None:
            return ("", "")

        root = tree.getroot()
        author = self._get_text(root, f"{{{DC_NS}}}creator")
        last_mod = self._get_text(root, f"{{{CP_NS}}}lastModifiedBy")
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
            if CORE_XML_PATH not in names:
                return
            core_bytes = zr.read(CORE_XML_PATH)

        tree = etree.fromstring(core_bytes)

        if author is not None:
            self._set_or_remove(tree, f"{{{DC_NS}}}creator", author)

        if last_modified_by is not None:
            self._set_or_remove(
                tree, f"{{{CP_NS}}}lastModifiedBy", last_modified_by
            )

        new_core = etree.tostring(tree, xml_declaration=True, encoding="UTF-8", standalone=True)

        out = BytesIO()
        with zipfile.ZipFile(buf, "r") as zr:
            with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as zw:
                for item in zr.infolist():
                    if item.filename == CORE_XML_PATH:
                        zw.writestr(item, new_core)
                    else:
                        zw.writestr(item, zr.read(item.filename))

        path.write_bytes(out.getvalue())

    def _read_core_xml(self, path: Path) -> etree._ElementTree | None:
        with zipfile.ZipFile(str(path), "r") as zf:
            if CORE_XML_PATH not in zf.namelist():
                return None
            data = zf.read(CORE_XML_PATH)
            return etree.ElementTree(etree.fromstring(data))

    def _get_text(self, root: etree._Element, tag: str) -> str:
        elem = root.find(tag)
        if elem is not None and elem.text:
            return elem.text.strip()
        return ""

    def _set_or_remove(
        self, root: etree._Element, tag: str, value: str
    ) -> None:
        elem = root.find(tag)
        if value == "":
            if elem is not None:
                root.remove(elem)
        else:
            if elem is None:
                elem = etree.SubElement(root, tag)
            elem.text = value
