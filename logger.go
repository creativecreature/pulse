package pulse

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// New wraps the construction of a charmbracelet logger
// in order to achieve coherent styles and settings.
func NewLogger() *log.Logger {
	logger := log.New(os.Stdout)
	logger.SetColorProfile(0)
	logger.SetLevel(log.DebugLevel)
	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Bold(true).
		MaxWidth(5).
		Foreground(lipgloss.Color("63"))

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		MaxWidth(5).
		Foreground(lipgloss.Color("204"))

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("FATAL").
		Bold(true).
		MaxWidth(5).
		Foreground(lipgloss.Color("134"))

	logger.SetStyles(styles)

	return logger
}
