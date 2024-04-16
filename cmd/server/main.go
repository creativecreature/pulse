package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse/disk"
	"github.com/creativecreature/pulse/server"
)

// ldflags.
var (
	serverName string
	port       string
)

func createLogger() *log.Logger {
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

func main() {
	logger := createLogger()

	diskStorage, err := disk.NewStorage()
	if err != nil {
		logger.Fatal(err, nil)
	}

	server, err := server.New(
		serverName,
		server.WithLog(logger),
		server.WithStorage(diskStorage),
	)
	if err != nil {
		logger.Fatal(err, nil)
	}

	err = server.Start(port)
	if err != nil {
		logger.Fatal(err, nil)
	}
}
