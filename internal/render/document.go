package render

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

// DocumentFromZIP extracts content from a DART document ZIP and
// converts it to markdown for terminal rendering.
func DocumentFromZIP(zipBytes []byte, rceptNo string) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return "", fmt.Errorf("ZIP 파싱 실패: %w", err)
	}

	type candidate struct {
		name string
		size uint64
		f    *zip.File
	}
	var candidates []candidate
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".xml" || ext == ".htm" || ext == ".html" {
			candidates = append(candidates, candidate{f.Name, f.UncompressedSize64, f})
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("문서 내에 읽을 수 있는 파일이 없습니다")
	}

	sort.Slice(candidates, func(i, j int) bool {
		iMatch := strings.Contains(candidates[i].name, rceptNo)
		jMatch := strings.Contains(candidates[j].name, rceptNo)
		if iMatch != jMatch {
			return iMatch
		}
		return candidates[i].size > candidates[j].size
	})

	var sb strings.Builder
	for i, c := range candidates {
		if i >= 3 {
			break
		}
		rc, err := c.f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		md := strings.TrimSpace(dartXMLToMarkdown(data))
		if md == "" {
			continue
		}
		if len(candidates) > 1 {
			fmt.Fprintf(&sb, "\n---\n*%s*\n\n", filepath.Base(c.name))
		}
		sb.WriteString(md)
		sb.WriteString("\n")
	}

	if sb.Len() == 0 {
		return "", fmt.Errorf("문서 내용을 추출할 수 없습니다")
	}
	return sb.String(), nil
}

// ── DART XML → Markdown ──────────────────────────────────────────────────────

// sanitizeXML replaces bare '<' characters that are not valid XML tag starts
// with '&lt;'. DART documents often contain Korean text in angle brackets
// used as visual emphasis (e.g. <이사ㆍ감사 보수현황>, < TV 시장점유율 >)
// which cause Go's xml.Decoder to fatal-error and stop parsing.
func sanitizeXML(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] != '<' {
			out = append(out, data[i])
			continue
		}
		// '<' — check next byte for valid XML name start or special marker
		if i+1 < len(data) {
			next := data[i+1]
			// Valid XML tag starters: a-z A-Z _ : / ! ? >
			if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') ||
				next == '_' || next == ':' || next == '/' || next == '!' ||
				next == '?' || next == '>' {
				out = append(out, '<')
			} else {
				out = append(out, '&', 'l', 't', ';')
			}
		} else {
			out = append(out, '<')
		}
	}
	return out
}

func dartXMLToMarkdown(data []byte) string {
	data = sanitizeXML(data)
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity

	p := &dartParser{dec: dec}
	p.run()
	return p.out.String()
}

// dartParser walks DART's XML token stream and emits markdown.
// It tracks an element-name stack so TITLE can know its heading depth.
type dartParser struct {
	dec   *xml.Decoder
	out   strings.Builder
	stack []string // element names currently open
}

// run is the main token loop.
func (p *dartParser) run() {
	for {
		tok, err := p.dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := upcase(t.Name.Local)
			p.stack = append(p.stack, name)
			endConsumed := p.handleStart(name)
			if endConsumed {
				p.pop()
			}
		case xml.EndElement:
			p.pop()
		}
		// CharData at the top level is ignored; text is gathered inside handlers.
	}
}

func (p *dartParser) pop() {
	if len(p.stack) > 0 {
		p.stack = p.stack[:len(p.stack)-1]
	}
}

// handleStart dispatches on element name.
// Returns true if the handler consumed the element's end tag (via collectText / parseTable / skipTo).
func (p *dartParser) handleStart(name string) (endConsumed bool) {
	switch {

	// ── skip: metadata & layout ──────────────────────────────────────────────
	case name == "SUMMARY" || name == "FORMULA-VERSION" ||
		name == "COLGROUP" || name == "COL" ||
		name == "IMAGE" || name == "IMG":
		p.skipTo(name)
		return true

	// ── transparent containers ───────────────────────────────────────────────
	// SECTION-*, BODY, COVER, TABLE-GROUP, TABLE structure elements, SPAN, LIBRARY
	// We let the token loop handle their children.
	case name == "DOCUMENT" || name == "BODY" || name == "COVER" ||
		name == "TABLE-GROUP" || name == "TBODY" || name == "THEAD" ||
		name == "SPAN" || name == "LIBRARY" || strings.HasPrefix(name, "SECTION-"):
		return false

	// ── headings ─────────────────────────────────────────────────────────────
	case name == "DOCUMENT-NAME":
		if text := p.collectText(); text != "" {
			fmt.Fprintf(&p.out, "# %s\n\n", text)
		}
		return true

	case name == "COMPANY-NAME":
		if text := p.collectText(); text != "" {
			fmt.Fprintf(&p.out, "> %s\n\n", text)
		}
		return true

	case name == "COVER-TITLE":
		if text := collapse(p.collectText()); text != "" {
			fmt.Fprintf(&p.out, "# %s\n\n", text)
		}
		return true

	case name == "TITLE":
		if text := collapse(p.collectText()); text != "" {
			level := p.titleLevel()
			fmt.Fprintf(&p.out, "%s %s\n\n", hashes(level), text)
		}
		return true

	// ── IMG-CAPTION ──────────────────────────────────────────────────────────
	case name == "IMG-CAPTION":
		if text := collapse(p.collectText()); text != "" {
			fmt.Fprintf(&p.out, "*%s*\n\n", text)
		}
		return true

	// ── paragraph ────────────────────────────────────────────────────────────
	case name == "P":
		text := strings.TrimSpace(p.collectText())
		if text != "" {
			fmt.Fprintf(&p.out, "%s\n\n", text)
		}
		return true

	// ── page break ───────────────────────────────────────────────────────────
	case name == "PGBRK":
		// self-contained element (no content), end tag handled by the token loop
		return false

	// ── table ────────────────────────────────────────────────────────────────
	case name == "TABLE":
		p.parseTable()
		return true

	// ── ignore TR/TD at top level (shouldn't appear) ─────────────────────────
	default:
		return false
	}
}

// titleLevel returns the markdown heading level for a <TITLE> element
// based on the deepest SECTION-N currently on the stack.
func (p *dartParser) titleLevel() int {
	for i := len(p.stack) - 1; i >= 0; i-- {
		s := p.stack[i]
		if strings.HasPrefix(s, "SECTION-") {
			// SECTION-1 → ##(2), SECTION-2 → ###(3), SECTION-3 → ####(4), ...
			var n int
			fmt.Sscanf(s[len("SECTION-"):], "%d", &n)
			if n > 0 {
				return n + 1
			}
		}
	}
	return 2 // default: ##
}

// collectText reads tokens until the matching end element and returns
// concatenated text content, preserving inline whitespace.
func (p *dartParser) collectText() string {
	var buf strings.Builder
	depth := 1
	for depth > 0 {
		tok, err := p.dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			buf.WriteString(string(t))
		}
	}
	return buf.String()
}

// skipTo consumes all tokens until the matching end element for name.
func (p *dartParser) skipTo(name string) {
	depth := 1
	for depth > 0 {
		tok, err := p.dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if upcase(t.Name.Local) == name {
				depth++
			}
		case xml.EndElement:
			if upcase(t.Name.Local) == name {
				depth--
			}
		}
	}
}

// parseTable collects TABLE content and writes a markdown table.
func (p *dartParser) parseTable() {
	var rows [][]string
	var curRow []string
	depth := 1

	for depth > 0 {
		tok, err := p.dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			elem := upcase(t.Name.Local)
			switch elem {
			case "TABLE":
				depth++
			case "TR":
				curRow = nil
			case "TD", "TH", "TU", "TE":
				// Collect cell text (may contain nested P, SPAN, etc.)
				text := p.collectText()
				text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
				text = strings.Join(strings.Fields(text), " ")
				// Escape pipes inside cells
				text = strings.ReplaceAll(text, "|", "｜")
				curRow = append(curRow, text)
			// column layout: skip (no content)
			case "COL", "COLGROUP":
				p.skipTo(elem)
			// THEAD/TBODY are transparent structural wrappers; TR/TD inside will be handled
			case "THEAD", "TBODY":
				// transparent — do nothing
			// embedded content: skip
			case "IMAGE", "IMG", "LIBRARY", "SUMMARY":
				p.skipTo(elem)
			}
		case xml.EndElement:
			elem := upcase(t.Name.Local)
			switch elem {
			case "TABLE":
				depth--
			case "TR":
				if hasContent(curRow) {
					rows = append(rows, curRow)
				}
				curRow = nil
			}
		}
	}

	if len(rows) == 0 {
		return
	}

	// If every cell in the table is empty or whitespace, skip it.
	nonEmpty := 0
	for _, r := range rows {
		for _, c := range r {
			if c != "" {
				nonEmpty++
			}
		}
	}
	if nonEmpty == 0 {
		return
	}

	// Single-cell, single-row → render as paragraph.
	if len(rows) == 1 && len(rows[0]) == 1 {
		fmt.Fprintf(&p.out, "%s\n\n", rows[0][0])
		return
	}

	// Find max column count.
	maxCols := 0
	for _, r := range rows {
		if len(r) > maxCols {
			maxCols = len(r)
		}
	}

	writeRow := func(cells []string) {
		p.out.WriteString("|")
		for i := 0; i < maxCols; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			fmt.Fprintf(&p.out, " %s |", cell)
		}
		p.out.WriteString("\n")
	}

	writeRow(rows[0])
	p.out.WriteString("|")
	for i := 0; i < maxCols; i++ {
		p.out.WriteString("---|")
	}
	p.out.WriteString("\n")
	for _, r := range rows[1:] {
		writeRow(r)
	}
	p.out.WriteString("\n")
}

// ── helpers ──────────────────────────────────────────────────────────────────

func upcase(s string) string { return strings.ToUpper(s) }

// collapse normalises runs of whitespace into a single space.
func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

// hashes returns n '#' characters.
func hashes(n int) string {
	if n < 1 {
		n = 1
	}
	if n > 6 {
		n = 6
	}
	return strings.Repeat("#", n)
}

// hasContent reports whether any cell in row has non-empty text.
func hasContent(row []string) bool {
	for _, c := range row {
		if c != "" {
			return true
		}
	}
	return false
}
