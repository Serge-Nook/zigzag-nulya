"""Модуль изменения временных штампов файловой системы."""

from __future__ import annotations

import os
import platform
import subprocess
from datetime import datetime, timedelta
from pathlib import Path

from topor.core.models import DateMode, TimeSettings


def apply_timestamps(path: Path, settings: TimeSettings) -> None:
    """Применить настройки временных штампов к файлу."""
    stat = path.stat()
    current_ctime = datetime.fromtimestamp(stat.st_ctime)
    current_mtime = datetime.fromtimestamp(stat.st_mtime)

    new_ctime = _resolve_time(
        settings.creation_mode,
        current_ctime,
        current_mtime,
        settings.creation_absolute,
        settings.creation_offset,
    )
    new_mtime = _resolve_time(
        settings.modification_mode,
        current_mtime,
        current_ctime,
        settings.modification_absolute,
        settings.modification_offset,
    )

    if new_mtime is not None:
        _set_modification_time(path, new_mtime)

    if new_ctime is not None:
        _set_creation_time(path, new_ctime)


def _resolve_time(
    mode: DateMode,
    current: datetime,
    other: datetime,
    absolute: datetime | None,
    offset: timedelta,
) -> datetime | None:
    """Вычислить новое время в зависимости от режима."""
    if mode == DateMode.NONE:
        return None
    if mode == DateMode.ABSOLUTE:
        return absolute
    if mode == DateMode.OFFSET:
        return current + offset
    if mode == DateMode.EQUAL_CREATE_TO_MODIFY:
        return other
    if mode == DateMode.EQUAL_MODIFY_TO_CREATE:
        return other
    return None


def _set_modification_time(path: Path, dt: datetime) -> None:
    """Установить время последнего изменения файла."""
    ts = dt.timestamp()
    atime = path.stat().st_atime
    os.utime(str(path), (atime, ts))


def _set_creation_time(path: Path, dt: datetime) -> None:
    """Установить время создания файла (платформозависимо)."""
    system = platform.system()
    ts = dt.timestamp()

    if system == "Windows":
        _set_creation_time_windows(path, dt)
    elif system == "Darwin":
        _set_creation_time_macos(path, dt)
    else:
        _set_creation_time_linux(path, ts)


def _set_creation_time_windows(path: Path, dt: datetime) -> None:
    """Установить время создания на Windows через win32."""
    try:
        import pywintypes
        import win32file

        ts = pywintypes.Time(dt)
        handle = win32file.CreateFile(
            str(path),
            win32file.GENERIC_WRITE,
            0,
            None,
            win32file.OPEN_EXISTING,
            0,
            None,
        )
        win32file.SetFileTime(handle, ts, None, None)
        handle.Close()
    except ImportError:
        pass


def _set_creation_time_macos(path: Path, dt: datetime) -> None:
    """Установить время создания на macOS через SetFile."""
    formatted = dt.strftime("%m/%d/%Y %H:%M:%S")
    try:
        subprocess.run(
            ["SetFile", "-d", formatted, str(path)],
            check=True,
            capture_output=True,
        )
    except (subprocess.CalledProcessError, FileNotFoundError):
        pass


def _set_creation_time_linux(path: Path, ts: float) -> None:
    """На Linux нет стандартного способа; обновляем mtime и atime."""
    try:
        os.utime(str(path), (ts, path.stat().st_mtime))
    except OSError:
        pass
