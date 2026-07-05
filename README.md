# Топор — Массовый редактор метаданных и временных штампов файлов

**Автор:** Горшков Сергей Владимирович  
**Сайт:** [nookbat.ru](https://nookbat.ru/)  
**Лицензия:** MIT

## Описание

Кроссплатформенное (Windows / Linux / macOS) десктопное приложение для пакетного изменения атрибутов файлов офисных и текстовых форматов.

### Возможности

- Изменение даты создания и даты последнего изменения файлов
- Изменение метаданных автора (Author) и последнего редактора (Last Modified By)
- Поддержка множества форматов: `docx`, `xlsx`, `docm`, `xlsm`, `doc`, `xls`, `odt`, `html`, `rtf`, `txt`, `xml`
- Три режима установки дат: абсолютная дата, сдвиг, приравнивание
- Drag & Drop файлов и папок
- Рекурсивный обход каталогов
- Фильтрация по маске
- Создание резервных копий (.bak)
- Журнал операций с подробными сообщениями
- Многопоточная обработка с прогресс-баром и кнопкой отмены
- Русский интерфейс

## Поддерживаемые форматы

| Группа | Форматы | Механика обработки |
|--------|---------|-------------------|
| Office Open XML | docx, xlsx, docm, xlsm | ZIP-контейнеры (Core Properties) |
| OLE / Старые форматы | doc, xls | Бинарные структуры (OLE Document Summary) |
| OpenDocument | odt | ZIP-контейнеры (meta.xml) |
| HTML | html, htm | Мета-теги `<meta name="author">` |
| RTF | rtf | Внутренние тэги `{\author ...}` |
| Текст / Разметка | txt, xml | Файловая система + расширенные атрибуты |

## Установка

### Из исходников

```bash
# Требования: Python 3.10+
pip install -r requirements.txt

# Запуск
python -m topor
```

### Установка как пакет

```bash
pip install .
topor
```

## Сборка установочных пакетов

### Windows (.exe)

Требуется [Inno Setup](https://jrsoftware.org/isinfo.php):

```bash
pip install pyinstaller
pyinstaller --noconfirm --onedir --name topor --windowed src/topor/__main__.py
# Затем скомпилировать packaging/windows/topor.iss через Inno Setup
```

### Linux (.deb)

```bash
./packaging/linux/build_deb.sh
```

### Linux (AppImage)

```bash
./packaging/linux/build_appimage.sh
```

### macOS (.dmg)

```bash
./packaging/macos/build_dmg.sh
```

## Структура проекта

```
src/topor/
├── app.py                  # Точка входа
├── core/
│   ├── engine.py           # Движок обработки
│   ├── file_discovery.py   # Поиск файлов
│   ├── models.py           # Модели данных
│   ├── timestamps.py       # Изменение временных штампов
│   └── handlers/
│       ├── base.py         # Базовый обработчик
│       ├── ooxml.py        # docx, xlsx, docm, xlsm
│       ├── ole.py          # doc, xls
│       ├── odt.py          # odt
│       ├── html_handler.py # html
│       ├── rtf.py          # rtf
│       └── plaintext.py    # txt, xml
├── gui/
│   ├── main_window.py      # Главное окно
│   ├── file_table.py       # Таблица предпросмотра
│   ├── time_panel.py       # Панель временных штампов
│   ├── author_panel.py     # Панель автора
│   ├── log_panel.py        # Журнал операций
│   └── styles.py           # Стили интерфейса
└── utils/
    └── backup.py           # Создание резервных копий
```

## Системные требования

- Python 3.10 или выше
- PyQt6
- Поддерживаемые ОС: Windows 10/11, Ubuntu 20.04+, macOS 10.15+

## Лицензия

MIT License — см. файл [LICENSE](LICENSE).
