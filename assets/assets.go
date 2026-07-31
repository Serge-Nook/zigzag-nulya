// Package assets содержит встроенные в исполняемый файл ресурсы программы.
package assets

import _ "embed"

// FontRegular — шрифт с поддержкой кириллицы для формирования PDF.
//
//go:embed fonts/DejaVuSans.ttf
var FontRegular []byte

// FontBold — жирное начертание шрифта для формирования PDF.
//
//go:embed fonts/DejaVuSans-Bold.ttf
var FontBold []byte
