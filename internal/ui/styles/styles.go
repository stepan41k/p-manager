package styles

import "github.com/charmbracelet/lipgloss"

var (
    MainColor = lipgloss.Color("#7D56F4")
    ErrorColor = lipgloss.Color("#FF0000")
    
    TitleStyle = lipgloss.NewStyle().
        Background(MainColor).
        Foreground(lipgloss.Color("230")).
        Padding(0, 1)

    BoxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(MainColor).
        Padding(1)
)