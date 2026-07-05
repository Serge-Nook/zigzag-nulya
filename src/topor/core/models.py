"""Модели данных приложения Топор."""

from __future__ import annotations

import enum
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from pathlib import Path


class DateMode(enum.Enum):
    """Режим установки дат."""

    NONE = "none"
    ABSOLUTE = "absolute"
    OFFSET = "offset"
    EQUAL_CREATE_TO_MODIFY = "equal_create_to_modify"
    EQUAL_MODIFY_TO_CREATE = "equal_modify_to_create"


class AuthorAction(enum.Enum):
    """Действие над полями автора."""

    NONE = "none"
    DELETE = "delete"
    OVERWRITE = "overwrite"


class ProcessingStatus(enum.Enum):
    """Статус обработки файла."""

    PENDING = "pending"
    SUCCESS = "success"
    SKIPPED = "skipped"
    ERROR = "error"


@dataclass
class FileMetadata:
    """Текущие метаданные файла."""

    path: Path
    size: int = 0
    creation_time: datetime | None = None
    modification_time: datetime | None = None
    author: str = ""
    last_modified_by: str = ""
    format_group: str = ""

    @property
    def extension(self) -> str:
        return self.path.suffix.lower()

    @property
    def filename(self) -> str:
        return self.path.name

    @property
    def size_display(self) -> str:
        if self.size < 1024:
            return f"{self.size} Б"
        elif self.size < 1024 * 1024:
            return f"{self.size / 1024:.1f} КБ"
        elif self.size < 1024 * 1024 * 1024:
            return f"{self.size / (1024 * 1024):.1f} МБ"
        else:
            return f"{self.size / (1024 * 1024 * 1024):.1f} ГБ"


@dataclass
class TimeSettings:
    """Настройки изменения временных штампов."""

    creation_mode: DateMode = DateMode.NONE
    modification_mode: DateMode = DateMode.NONE
    creation_absolute: datetime | None = None
    modification_absolute: datetime | None = None
    creation_offset: timedelta = field(default_factory=timedelta)
    modification_offset: timedelta = field(default_factory=timedelta)


@dataclass
class AuthorSettings:
    """Настройки изменения метаданных автора."""

    author_action: AuthorAction = AuthorAction.NONE
    last_modified_by_action: AuthorAction = AuthorAction.NONE
    new_author: str = ""
    new_last_modified_by: str = ""


@dataclass
class ProcessingSettings:
    """Общие настройки обработки."""

    time_settings: TimeSettings = field(default_factory=TimeSettings)
    author_settings: AuthorSettings = field(default_factory=AuthorSettings)
    create_backup: bool = False
    include_subdirectories: bool = False
    format_filter: set[str] = field(default_factory=set)


@dataclass
class ProcessingResult:
    """Результат обработки одного файла."""

    path: Path
    status: ProcessingStatus = ProcessingStatus.PENDING
    message: str = ""
