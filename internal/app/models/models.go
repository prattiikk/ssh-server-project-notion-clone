package models

import (
	"github.com/charmbracelet/lipgloss"

	_ "github.com/lib/pq"
)

// Define the list item view model struct
type ListItemViewModel struct {
	ItemTitle       string
	Desc            string
	Content         string
	ShowItemContent bool
}
type Dimensions struct {
	TotalWidth  int
	TotalHeight int
}

// Methods to fulfill the list.Item interface
func (i ListItemViewModel) FilterValue() string { return i.ItemTitle }
func (i ListItemViewModel) Title() string       { return i.ItemTitle }
func (i ListItemViewModel) Description() string { return i.Desc }

// Renders the individual item view
func (m ListItemViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Margin(4, 10, 0).Height(16).Width(100).Border(lipgloss.NormalBorder(), false, false, true, false).Render(m.Content),
		lipgloss.NewStyle().Height(2).MarginLeft(10).MarginTop(2).Render("ctrl+a: exit alt screen"),
	)
}

// Struct to hold a slice of items
type ItemsMsg struct {
	Items []ListItemViewModel
}
