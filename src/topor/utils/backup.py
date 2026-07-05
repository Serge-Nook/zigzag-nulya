"""Утилита создания резервных копий."""

from __future__ import annotations

import shutil
from pathlib import Path


def create_backup(path: Path) -> Path:
    """Создать .bak копию файла. Возвращает путь к копии."""
    backup_path = path.with_suffix(path.suffix + ".bak")
    counter = 1
    while backup_path.exists():
        backup_path = path.with_suffix(f"{path.suffix}.bak{counter}")
        counter += 1
    shutil.copy2(str(path), str(backup_path))
    return backup_path
