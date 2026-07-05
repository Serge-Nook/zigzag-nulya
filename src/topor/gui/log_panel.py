"""Панель лога операций."""

from __future__ import annotations

from datetime import datetime

from PyQt6.QtWidgets import (
    QHBoxLayout,
    QPushButton,
    QTextEdit,
    QVBoxLayout,
    QWidget,
)


class LogPanel(QWidget):
    """Виджет лога с кнопкой очистки."""

    def __init__(self, parent: QWidget | None = None) -> None:
        super().__init__(parent)
        layout = QVBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)

        self._log = QTextEdit()
        self._log.setObjectName("logWidget")
        self._log.setReadOnly(True)
        layout.addWidget(self._log)

        btn_row = QHBoxLayout()
        btn_row.addStretch()
        self._clear_btn = QPushButton("Очистить лог")
        self._clear_btn.setObjectName("secondaryButton")
        self._clear_btn.setFixedWidth(120)
        self._clear_btn.clicked.connect(self.clear)
        btn_row.addWidget(self._clear_btn)
        layout.addLayout(btn_row)

    def append(self, message: str) -> None:
        """Добавить запись в лог."""
        timestamp = datetime.now().strftime("%H:%M:%S")
        self._log.append(f"[{timestamp}] {message}")

    def clear(self) -> None:
        """Очистить лог."""
        self._log.clear()

    def get_text(self) -> str:
        """Получить весь текст лога."""
        return self._log.toPlainText()
