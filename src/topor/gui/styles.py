"""Стили и константы GUI."""

APP_STYLESHEET = """
QMainWindow {
    background-color: #f5f5f5;
}

QGroupBox {
    font-weight: bold;
    border: 1px solid #cccccc;
    border-radius: 4px;
    margin-top: 12px;
    padding-top: 16px;
}

QGroupBox::title {
    subcontrol-origin: margin;
    left: 10px;
    padding: 0 5px;
}

QTableWidget {
    background-color: white;
    alternate-background-color: #f9f9f9;
    gridline-color: #e0e0e0;
    selection-background-color: #0078d4;
    selection-color: white;
    border: 1px solid #cccccc;
    border-radius: 2px;
}

QTableWidget::item {
    padding: 4px;
}

QHeaderView::section {
    background-color: #e8e8e8;
    padding: 4px 8px;
    border: none;
    border-right: 1px solid #cccccc;
    border-bottom: 1px solid #cccccc;
    font-weight: bold;
}

QPushButton {
    background-color: #0078d4;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 6px 16px;
    min-height: 24px;
    font-size: 13px;
}

QPushButton:hover {
    background-color: #106ebe;
}

QPushButton:pressed {
    background-color: #005a9e;
}

QPushButton:disabled {
    background-color: #cccccc;
    color: #888888;
}

QPushButton#dangerButton {
    background-color: #d32f2f;
}

QPushButton#dangerButton:hover {
    background-color: #b71c1c;
}

QPushButton#secondaryButton {
    background-color: #757575;
}

QPushButton#secondaryButton:hover {
    background-color: #616161;
}

QTextEdit#logWidget {
    background-color: #1e1e1e;
    color: #cccccc;
    font-family: "Consolas", "Courier New", monospace;
    font-size: 12px;
    border: 1px solid #cccccc;
    border-radius: 2px;
}

QProgressBar {
    border: 1px solid #cccccc;
    border-radius: 4px;
    text-align: center;
    background-color: #e0e0e0;
    height: 22px;
}

QProgressBar::chunk {
    background-color: #0078d4;
    border-radius: 3px;
}

QDateTimeEdit, QSpinBox, QLineEdit, QComboBox {
    padding: 4px 8px;
    border: 1px solid #cccccc;
    border-radius: 3px;
    background-color: white;
    min-height: 22px;
}

QCheckBox, QRadioButton {
    spacing: 6px;
}

QTabWidget::pane {
    border: 1px solid #cccccc;
    border-radius: 2px;
    background-color: white;
}

QTabBar::tab {
    background-color: #e8e8e8;
    padding: 6px 16px;
    margin-right: 2px;
    border: 1px solid #cccccc;
    border-bottom: none;
    border-top-left-radius: 4px;
    border-top-right-radius: 4px;
}

QTabBar::tab:selected {
    background-color: white;
    border-bottom: 1px solid white;
}

QSplitter::handle {
    background-color: #cccccc;
    height: 3px;
}

QStatusBar {
    background-color: #e8e8e8;
    border-top: 1px solid #cccccc;
}
"""
