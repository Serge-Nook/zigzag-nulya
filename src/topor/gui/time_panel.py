"""Панель управления временными штампами."""

from __future__ import annotations

from datetime import datetime, timedelta

from PyQt6.QtCore import QDateTime
from PyQt6.QtWidgets import (
    QButtonGroup,
    QDateTimeEdit,
    QGridLayout,
    QGroupBox,
    QHBoxLayout,
    QLabel,
    QRadioButton,
    QSpinBox,
    QVBoxLayout,
    QWidget,
)

from topor.core.models import DateMode, TimeSettings


class TimePanel(QWidget):
    """Панель настроек временных штампов."""

    def __init__(self, parent: QWidget | None = None) -> None:
        super().__init__(parent)
        layout = QVBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)

        # --- Дата создания ---
        self._create_group = QGroupBox("Дата создания (Creation Time)")
        create_layout = QVBoxLayout(self._create_group)
        self._create_mode_group = QButtonGroup(self)

        self._create_none = QRadioButton("Не изменять")
        self._create_none.setChecked(True)
        self._create_absolute = QRadioButton("Установить абсолютную дату")
        self._create_offset = QRadioButton("Сдвиг относительно текущей")
        self._create_equal = QRadioButton("Приравнять к дате изменения")

        self._create_mode_group.addButton(self._create_none, 0)
        self._create_mode_group.addButton(self._create_absolute, 1)
        self._create_mode_group.addButton(self._create_offset, 2)
        self._create_mode_group.addButton(self._create_equal, 3)

        create_layout.addWidget(self._create_none)

        abs_row = QHBoxLayout()
        abs_row.addWidget(self._create_absolute)
        self._create_dt = QDateTimeEdit()
        self._create_dt.setCalendarPopup(True)
        self._create_dt.setDateTime(QDateTime.currentDateTime())
        self._create_dt.setDisplayFormat("dd.MM.yyyy HH:mm:ss")
        self._create_dt.setEnabled(False)
        abs_row.addWidget(self._create_dt)
        create_layout.addLayout(abs_row)

        offset_row = QHBoxLayout()
        offset_row.addWidget(self._create_offset)
        self._create_offset_days = QSpinBox()
        self._create_offset_days.setRange(-9999, 9999)
        self._create_offset_days.setSuffix(" дн.")
        self._create_offset_days.setEnabled(False)
        offset_row.addWidget(self._create_offset_days)
        self._create_offset_hours = QSpinBox()
        self._create_offset_hours.setRange(-23, 23)
        self._create_offset_hours.setSuffix(" ч.")
        self._create_offset_hours.setEnabled(False)
        offset_row.addWidget(self._create_offset_hours)
        self._create_offset_mins = QSpinBox()
        self._create_offset_mins.setRange(-59, 59)
        self._create_offset_mins.setSuffix(" мин.")
        self._create_offset_mins.setEnabled(False)
        offset_row.addWidget(self._create_offset_mins)
        create_layout.addLayout(offset_row)

        create_layout.addWidget(self._create_equal)
        layout.addWidget(self._create_group)

        # --- Дата изменения ---
        self._modify_group = QGroupBox("Дата изменения (Modification Time)")
        modify_layout = QVBoxLayout(self._modify_group)
        self._modify_mode_group = QButtonGroup(self)

        self._modify_none = QRadioButton("Не изменять")
        self._modify_none.setChecked(True)
        self._modify_absolute = QRadioButton("Установить абсолютную дату")
        self._modify_offset = QRadioButton("Сдвиг относительно текущей")
        self._modify_equal = QRadioButton("Приравнять к дате создания")

        self._modify_mode_group.addButton(self._modify_none, 0)
        self._modify_mode_group.addButton(self._modify_absolute, 1)
        self._modify_mode_group.addButton(self._modify_offset, 2)
        self._modify_mode_group.addButton(self._modify_equal, 3)

        modify_layout.addWidget(self._modify_none)

        abs_row2 = QHBoxLayout()
        abs_row2.addWidget(self._modify_absolute)
        self._modify_dt = QDateTimeEdit()
        self._modify_dt.setCalendarPopup(True)
        self._modify_dt.setDateTime(QDateTime.currentDateTime())
        self._modify_dt.setDisplayFormat("dd.MM.yyyy HH:mm:ss")
        self._modify_dt.setEnabled(False)
        abs_row2.addWidget(self._modify_dt)
        modify_layout.addLayout(abs_row2)

        offset_row2 = QHBoxLayout()
        offset_row2.addWidget(self._modify_offset)
        self._modify_offset_days = QSpinBox()
        self._modify_offset_days.setRange(-9999, 9999)
        self._modify_offset_days.setSuffix(" дн.")
        self._modify_offset_days.setEnabled(False)
        offset_row2.addWidget(self._modify_offset_days)
        self._modify_offset_hours = QSpinBox()
        self._modify_offset_hours.setRange(-23, 23)
        self._modify_offset_hours.setSuffix(" ч.")
        self._modify_offset_hours.setEnabled(False)
        offset_row2.addWidget(self._modify_offset_hours)
        self._modify_offset_mins = QSpinBox()
        self._modify_offset_mins.setRange(-59, 59)
        self._modify_offset_mins.setSuffix(" мин.")
        self._modify_offset_mins.setEnabled(False)
        offset_row2.addWidget(self._modify_offset_mins)
        modify_layout.addLayout(offset_row2)

        modify_layout.addWidget(self._modify_equal)
        layout.addWidget(self._modify_group)

        layout.addStretch()

        # Сигналы
        self._create_mode_group.idToggled.connect(self._on_create_mode_changed)
        self._modify_mode_group.idToggled.connect(self._on_modify_mode_changed)

    def _on_create_mode_changed(self, button_id: int, checked: bool) -> None:
        if not checked:
            return
        self._create_dt.setEnabled(button_id == 1)
        is_offset = button_id == 2
        self._create_offset_days.setEnabled(is_offset)
        self._create_offset_hours.setEnabled(is_offset)
        self._create_offset_mins.setEnabled(is_offset)

    def _on_modify_mode_changed(self, button_id: int, checked: bool) -> None:
        if not checked:
            return
        self._modify_dt.setEnabled(button_id == 1)
        is_offset = button_id == 2
        self._modify_offset_days.setEnabled(is_offset)
        self._modify_offset_hours.setEnabled(is_offset)
        self._modify_offset_mins.setEnabled(is_offset)

    def get_settings(self) -> TimeSettings:
        """Собрать настройки из элементов управления."""
        settings = TimeSettings()

        create_id = self._create_mode_group.checkedId()
        if create_id == 0:
            settings.creation_mode = DateMode.NONE
        elif create_id == 1:
            settings.creation_mode = DateMode.ABSOLUTE
            settings.creation_absolute = self._create_dt.dateTime().toPyDateTime()
        elif create_id == 2:
            settings.creation_mode = DateMode.OFFSET
            settings.creation_offset = timedelta(
                days=self._create_offset_days.value(),
                hours=self._create_offset_hours.value(),
                minutes=self._create_offset_mins.value(),
            )
        elif create_id == 3:
            settings.creation_mode = DateMode.EQUAL_CREATE_TO_MODIFY

        modify_id = self._modify_mode_group.checkedId()
        if modify_id == 0:
            settings.modification_mode = DateMode.NONE
        elif modify_id == 1:
            settings.modification_mode = DateMode.ABSOLUTE
            settings.modification_absolute = self._modify_dt.dateTime().toPyDateTime()
        elif modify_id == 2:
            settings.modification_mode = DateMode.OFFSET
            settings.modification_offset = timedelta(
                days=self._modify_offset_days.value(),
                hours=self._modify_offset_hours.value(),
                minutes=self._modify_offset_mins.value(),
            )
        elif modify_id == 3:
            settings.modification_mode = DateMode.EQUAL_MODIFY_TO_CREATE

        return settings
