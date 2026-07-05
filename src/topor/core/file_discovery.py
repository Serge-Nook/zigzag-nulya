"""Модуль обнаружения и фильтрации файлов."""

from __future__ import annotations

import fnmatch
from pathlib import Path

from topor.core.handlers import HANDLER_MAP

SUPPORTED_EXTENSIONS = set(HANDLER_MAP.keys())


def discover_files(
    paths: list[Path],
    include_subdirectories: bool = False,
    format_filter: set[str] | None = None,
) -> list[Path]:
    """Собрать список файлов для обработки.

    Args:
        paths: Список путей (файлы и/или каталоги).
        include_subdirectories: Рекурсивный обход каталогов.
        format_filter: Набор расширений для фильтрации (например, {".docx", ".xlsx"}).
            Если None или пустое — все поддерживаемые форматы.
    """
    allowed = format_filter if format_filter else SUPPORTED_EXTENSIONS
    result: list[Path] = []

    for p in paths:
        if p.is_file():
            if p.suffix.lower() in allowed:
                result.append(p)
        elif p.is_dir():
            _scan_directory(p, allowed, include_subdirectories, result)

    return sorted(set(result))


def _scan_directory(
    directory: Path,
    allowed: set[str],
    recursive: bool,
    result: list[Path],
) -> None:
    """Сканировать каталог на предмет подходящих файлов."""
    try:
        entries = sorted(directory.iterdir())
    except PermissionError:
        return

    for entry in entries:
        if entry.is_file() and entry.suffix.lower() in allowed:
            result.append(entry)
        elif entry.is_dir() and recursive:
            _scan_directory(entry, allowed, recursive, result)


def filter_by_mask(files: list[Path], mask: str) -> list[Path]:
    """Фильтрация списка файлов по маске (например, '*.docx')."""
    masks = [m.strip() for m in mask.split(";") if m.strip()]
    if not masks:
        return files
    return [
        f
        for f in files
        if any(fnmatch.fnmatch(f.name, m) for m in masks)
    ]
