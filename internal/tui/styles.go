package tui

import "github.com/charmbracelet/lipgloss"

// Colour palette — follows ux-standards.md §11.
var (
	colorGreen  = lipgloss.Color("2")
	colorYellow = lipgloss.Color("11")
	colorRed    = lipgloss.Color("1")
	colorBlue   = lipgloss.Color("4")
	colorMuted  = lipgloss.Color("8")
)

var (
	styleCompleted = lipgloss.NewStyle().Foreground(colorGreen)
	stylePending   = lipgloss.NewStyle().Foreground(colorYellow)
	styleError     = lipgloss.NewStyle().Foreground(colorRed)
	stylePlaying   = lipgloss.NewStyle().Foreground(colorBlue)
	styleMuted     = lipgloss.NewStyle().Foreground(colorMuted)
	styleBold      = lipgloss.NewStyle().Bold(true)
	styleFooterSep = lipgloss.NewStyle().Foreground(colorMuted)
)

// footerDivider is the full-width separator above the help footer.
const footerDivider = "──────────────────────────────────────────────────────"

// formatFooter renders a minimal footer line from key–label pairs.
// e.g. formatFooter("space", "select", "i", "import") → "[space] select   [i] import"
func formatFooter(pairs ...string) string {
	var parts []string
	for i := 0; i+1 < len(pairs); i += 2 {
		key, label := pairs[i], pairs[i+1]
		parts = append(parts, styleMuted.Render("["+key+"]")+" "+label)
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "   "
		}
		result += p
	}
	return result
}
