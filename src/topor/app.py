"""Точка входа приложения Топор."""

from __future__ import annotations

import logging
import sys

from PyQt6.QtWidgets import QApplication

from topor import __version__
from topor.gui.main_window import MainWindow
from topor.gui.styles import APP_STYLESHEET


def setup_logging() -> None:
    """Настроить логирование."""
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
        datefmt="%H:%M:%S",
    )


def main() -> None:
    """Запустить приложение."""
    setup_logging()
    logger = logging.getLogger("topor")
    logger.info("Запуск Топор v%s", __version__)

    app = QApplication(sys.argv)
    app.setApplicationName("Топор")
    app.setApplicationVersion(__version__)
    app.setOrganizationName("Горшков С.В.")
    app.setStyleSheet(APP_STYLESHEET)

    window = MainWindow()
    window.show()

    sys.exit(app.exec())


if __name__ == "__main__":
    main()
