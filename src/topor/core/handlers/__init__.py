"""Обработчики форматов файлов."""

from topor.core.handlers.base import BaseHandler
from topor.core.handlers.ooxml import OOXMLHandler
from topor.core.handlers.ole import OLEHandler
from topor.core.handlers.odt import ODTHandler
from topor.core.handlers.html_handler import HTMLHandler
from topor.core.handlers.rtf import RTFHandler
from topor.core.handlers.plaintext import PlainTextHandler

HANDLER_MAP: dict[str, type[BaseHandler]] = {
    ".docx": OOXMLHandler,
    ".xlsx": OOXMLHandler,
    ".docm": OOXMLHandler,
    ".xlsm": OOXMLHandler,
    ".doc": OLEHandler,
    ".xls": OLEHandler,
    ".odt": ODTHandler,
    ".html": HTMLHandler,
    ".htm": HTMLHandler,
    ".rtf": RTFHandler,
    ".txt": PlainTextHandler,
    ".xml": PlainTextHandler,
}

__all__ = [
    "BaseHandler",
    "OOXMLHandler",
    "OLEHandler",
    "ODTHandler",
    "HTMLHandler",
    "RTFHandler",
    "PlainTextHandler",
    "HANDLER_MAP",
]
