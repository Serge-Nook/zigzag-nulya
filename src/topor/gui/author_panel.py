"""Панель управления метаданными автора."""

from __future__ import annotations

from PyQt6.QtWidgets import (
    QButtonGroup,
    QGroupBox,
    QHBoxLayout,
    QLabel,
    QLineEdit,
    QRadioButton,
    QVBoxLayout,
    QWidget,
)

from topor.core.models import AuthorAction, AuthorSettings


class AuthorPanel(QWidget):
    """Панель настроек метаданных автора."""

    def __init__(self, parent: QWidget | None = None) -> None:
        super().__init__(parent)
        layout = QVBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)

        # --- Автор ---
        self._author_group = QGroupBox("Автор (Author)")
        author_layout = QVBoxLayout(self._author_group)
        self._author_mode_group = QButtonGroup(self)

        self._author_none = QRadioButton("Не изменять")
        self._author_none.setChecked(True)
        self._author_delete = QRadioButton("Удалить")
        self._author_overwrite = QRadioButton("Перезаписать:")

        self._author_mode_group.addButton(self._author_none, 0)
        self._author_mode_group.addButton(self._author_delete, 1)
        self._author_mode_group.addButton(self._author_overwrite, 2)

        author_layout.addWidget(self._author_none)
        author_layout.addWidget(self._author_delete)

        overwrite_row = QHBoxLayout()
        overwrite_row.addWidget(self._author_overwrite)
        self._author_input = QLineEdit()
        self._author_input.setPlaceholderText("Новое имя автора")
        self._author_input.setEnabled(False)
        overwrite_row.addWidget(self._author_input)
        author_layout.addLayout(overwrite_row)

        layout.addWidget(self._author_group)

        # --- Последний редактор ---
        self._lmb_group = QGroupBox("Последний редактор (Last Modified By)")
        lmb_layout = QVBoxLayout(self._lmb_group)
        self._lmb_mode_group = QButtonGroup(self)

        self._lmb_none = QRadioButton("Не изменять")
        self._lmb_none.setChecked(True)
        self._lmb_delete = QRadioButton("Удалить")
        self._lmb_overwrite = QRadioButton("Перезаписать:")

        self._lmb_mode_group.addButton(self._lmb_none, 0)
        self._lmb_mode_group.addButton(self._lmb_delete, 1)
        self._lmb_mode_group.addButton(self._lmb_overwrite, 2)

        lmb_layout.addWidget(self._lmb_none)
        lmb_layout.addWidget(self._lmb_delete)

        lmb_overwrite_row = QHBoxLayout()
        lmb_overwrite_row.addWidget(self._lmb_overwrite)
        self._lmb_input = QLineEdit()
        self._lmb_input.setPlaceholderText("Новое имя последнего редактора")
        self._lmb_input.setEnabled(False)
        lmb_overwrite_row.addWidget(self._lmb_input)
        lmb_layout.addLayout(lmb_overwrite_row)

        layout.addWidget(self._lmb_group)
        layout.addStretch()

        # Сигналы
        self._author_mode_group.idToggled.connect(self._on_author_mode_changed)
        self._lmb_mode_group.idToggled.connect(self._on_lmb_mode_changed)

    def _on_author_mode_changed(self, button_id: int, checked: bool) -> None:
        if not checked:
            return
        self._author_input.setEnabled(button_id == 2)

    def _on_lmb_mode_changed(self, button_id: int, checked: bool) -> None:
        if not checked:
            return
        self._lmb_input.setEnabled(button_id == 2)

    def get_settings(self) -> AuthorSettings:
        """Собрать настройки из элементов управления."""
        settings = AuthorSettings()

        author_id = self._author_mode_group.checkedId()
        if author_id == 0:
            settings.author_action = AuthorAction.NONE
        elif author_id == 1:
            settings.author_action = AuthorAction.DELETE
        elif author_id == 2:
            settings.author_action = AuthorAction.OVERWRITE
            settings.new_author = self._author_input.text()

        lmb_id = self._lmb_mode_group.checkedId()
        if lmb_id == 0:
            settings.last_modified_by_action = AuthorAction.NONE
        elif lmb_id == 1:
            settings.last_modified_by_action = AuthorAction.DELETE
        elif lmb_id == 2:
            settings.last_modified_by_action = AuthorAction.OVERWRITE
            settings.new_last_modified_by = self._lmb_input.text()

        return settings
