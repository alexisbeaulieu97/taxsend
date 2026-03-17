package output

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Printer struct {
	logger  *log.Logger
	success lipgloss.Style
	warn    lipgloss.Style
}

func New(verbose bool) *Printer {
	level := log.InfoLevel
	if verbose {
		level = log.DebugLevel
	}
	lg := log.NewWithOptions(nil, log.Options{Level: level})
	return &Printer{logger: lg, success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")), warn: lipgloss.NewStyle().Foreground(lipgloss.Color("11"))}
}

func (p *Printer) Info(msg string, keyvals ...any)  { p.logger.Info(msg, keyvals...) }
func (p *Printer) Debug(msg string, keyvals ...any) { p.logger.Debug(msg, keyvals...) }
func (p *Printer) Error(msg string, keyvals ...any) { p.logger.Error(msg, keyvals...) }
func (p *Printer) Success(msg string)               { fmt.Println(p.success.Render(msg)) }
func (p *Printer) Warn(msg string)                  { fmt.Println(p.warn.Render(msg)) }
