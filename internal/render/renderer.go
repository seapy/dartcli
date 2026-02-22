package render

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// Renderer wraps glamour.TermRenderer with TTY detection.
type Renderer struct {
	style    string
	noColor  bool
}

// New creates a Renderer.
func New(style string, noColor bool) *Renderer {
	return &Renderer{style: style, noColor: noColor}
}

// Print renders markdown and writes it to stdout.
func (r *Renderer) Print(markdown string) error {
	return r.print(markdown, false)
}

// PrintWide renders markdown without a word-wrap width limit.
// Use this for content with wide tables (e.g. document view) to prevent
// glamour from truncating cell content with "...".
func (r *Renderer) PrintWide(markdown string) error {
	return r.print(markdown, true)
}

func (r *Renderer) print(markdown string, wide bool) error {
	style := r.resolveStyle()

	tableWrap := true
	opts := []glamour.TermRendererOption{
		glamour.WithStylePath(style),
		glamour.WithTableWrap(tableWrap),
	}

	if wide {
		// No fixed width → table columns expand to fit content, no truncation.
		opts = append(opts, glamour.WithWordWrap(0))
	} else {
		opts = append(opts, glamour.WithWordWrap(r.termWidth()))
	}

	renderer, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		// Fallback: plain output
		fmt.Println(markdown)
		return nil
	}

	out, err := renderer.Render(markdown)
	if err != nil {
		fmt.Println(markdown)
		return nil
	}
	fmt.Print(out)
	return nil
}

func (r *Renderer) resolveStyle() string {
	if r.noColor || !isatty.IsTerminal(os.Stdout.Fd()) {
		return "notty"
	}
	if r.style != "" && r.style != "auto" {
		return r.style
	}
	return "auto"
}

func (r *Renderer) termWidth() int {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return 120
	}
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	if w > 120 {
		return 120
	}
	return w
}
