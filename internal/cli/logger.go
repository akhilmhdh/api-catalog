package cli

import (
	"fmt"
	"strings"

	"github.com/1-platform/api-catalog/internal/cli/reportmanager"
	"github.com/charmbracelet/lipgloss"
)

type CliLogger struct{}

var infoTopic = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fc2ff"))
var warnTopic = lipgloss.NewStyle().Foreground(lipgloss.Color("#f0ab00"))
var errorTopic = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9190b"))
var successTopic = lipgloss.NewStyle().Foreground(lipgloss.Color("#5bc24e"))
var logTopic = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

var subject = lipgloss.NewStyle().PaddingLeft(2)
var divider = lipgloss.NewStyle().Foreground(lipgloss.Color("#bab8b8")).PaddingLeft(2).Render("-----------------------------------------")

var title = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fff5f")).PaddingLeft(15).Bold(true)

var reportTemplateTitle = lipgloss.NewStyle().PaddingLeft(2).Bold(true)
var reportTemplateValue = lipgloss.NewStyle().PaddingLeft(1)

var scoreCard = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true)

func NewCliLogger() *CliLogger {
	return &CliLogger{}
}

func (l *CliLogger) Info(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, infoTopic.Render("[ INFO ]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Log(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, logTopic.Render("[ LOG ]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Error(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, errorTopic.Render("[ ERROR ]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Warn(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, warnTopic.Render("[ WARN ]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Success(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, successTopic.Render("[ SUCCESS ]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Completed(info string) {
	block := lipgloss.JoinHorizontal(lipgloss.Top, successTopic.Render("[\u2713]"), subject.Render(info))
	fmt.Println(block)
}

func (l *CliLogger) Divider() {
	fmt.Println(divider)
}

func (l *CliLogger) Report(rule, method, path, message string) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s%s\n", reportTemplateTitle.Render("Rule:"), reportTemplateValue.Render(rule)))
	sb.WriteString(fmt.Sprintf("%s%s\n", reportTemplateTitle.Render("Method:"), reportTemplateValue.Render(method)))
	sb.WriteString(fmt.Sprintf("%s%s\n", reportTemplateTitle.Render("Path:"), reportTemplateValue.Render(path)))
	sb.WriteString(fmt.Sprintf("%s%s", reportTemplateTitle.Render("Message:"), reportTemplateValue.Render(message)))

	fmt.Println(sb.String())
}

func (l *CliLogger) Title(info string) {
	l.Divider()
	fmt.Println(title.Render(info))
	l.Divider()
}

func (l *CliLogger) ScoreCard(scores []reportmanager.Score) {
	var sb strings.Builder

	for _, score := range scores {

		sb.WriteString(fmt.Sprintf("%s:%s ::", reportTemplateTitle.Render("Category"), reportTemplateValue.Render(strings.Title(score.Category))))
		sb.WriteString(fmt.Sprintf("%s:%f\n", reportTemplateTitle.Render("Score"), score.Value))
	}

	fmt.Println(sb.String())
}
