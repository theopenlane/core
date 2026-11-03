// Package pdftext extracts the text layer of a pdf, decoding composite font strings through each font's ToUnicode map
//
// It exists because uploads are validated locally before any model call, and pdfcpu parses a pdf
// without extracting text. Reading the string operands directly is not enough: a report exported
// from Word uses subset fonts with Identity-H encoding, so its content streams carry two byte
// glyph ids rather than characters, and only the font's ToUnicode map turns those back into text
package pdftext
