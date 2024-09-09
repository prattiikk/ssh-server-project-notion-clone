package styles

import "github.com/charmbracelet/lipgloss"

// Styles
var ListStyle = lipgloss.NewStyle().Margin(4, 10, 0).Width(60)

var TextareaStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder(), false, true, false, false).
	// Padding(1, 2).
	MarginTop(2).
	BorderForeground(lipgloss.Color("#7571F9")).
	// Background(lipgloss.Color("#020d14")).
	Foreground(lipgloss.Color("#7571F9"))

var ViewportStyle = lipgloss.NewStyle().

	//Padding(1, 2).
	MarginTop(2).
	// BorderForeground(lipgloss.Color("63")).
	// Background(lipgloss.Color("#020d14")).
	Foreground(lipgloss.Color("#7571F9"))

var CenteredViewportStyle = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("#7571F9"))

// MarginLeft(40)

// var Style = lipgloss.NewStyle().
// 	Border(lipgloss.NormalBorder()).
// 	Padding(1, 2).
// 	BorderForeground(lipgloss.Color("#444444")).
// 	Foreground(lipgloss.Color("#7571F9"))

var FormStyle = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder()).
	Padding(1, 2).
	BorderForeground(lipgloss.Color("#7571F9")). // Purple border
	Foreground(lipgloss.Color("#7571F9")).       // Purple text
	Align(lipgloss.Center).                      // Center align content
	Width(30).                                   // Set a fixed width for the form
	Height(10)                                   // Set a fixed height for the form

var BannerText = `
██╗      ██████╗  ██████╗ ██╗███╗   ██╗
██║     ██╔═══██╗██╔════╝ ██║████╗  ██║
██║     ██║   ██║██║  ███╗██║██╔██╗ ██║
██║     ██║   ██║██║   ██║██║██║╚██╗██║
███████╗╚██████╔╝╚██████╔╝██║██║ ╚████║
╚══════╝ ╚═════╝  ╚═════╝ ╚═╝╚═╝  ╚═══╝
`

// logo style
var Logostyle = lipgloss.NewStyle().
	Width(50).
	Height(1).
	Align(lipgloss.Center).
	Foreground(lipgloss.Color("#7571F9")).
	Render("NotionTerm.sh")
