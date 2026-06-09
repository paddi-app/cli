package output

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

// colPad is the number of spaces separating adjacent columns.
const colPad = 2

// Table renders rows as a column-aligned table with a header row. Column
// widths are measured by terminal display width so that double-width runes
// (CJK, etc.) stay aligned.
func Table(w io.Writer, headers []string, rows [][]string) error {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = runewidth.StringWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				if cw := runewidth.StringWidth(cell); cw > widths[i] {
					widths[i] = cw
				}
			}
		}
	}

	var b strings.Builder
	writeRow(&b, headers, widths)
	for _, row := range rows {
		writeRow(&b, row, widths)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// writeRow writes one tab-free row, padding every cell but the last to its
// column's display width.
func writeRow(b *strings.Builder, cells []string, widths []int) {
	for i, width := range widths {
		var cell string
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(cell)
		if i < len(widths)-1 {
			b.WriteString(strings.Repeat(" ", width-runewidth.StringWidth(cell)+colPad))
		}
	}
	b.WriteByte('\n')
}

// JSON writes raw API JSON followed by a newline.
func JSON(w io.Writer, raw json.RawMessage) error {
	_, err := w.Write(append(bytes.TrimRight(raw, "\n"), '\n'))
	return err
}
