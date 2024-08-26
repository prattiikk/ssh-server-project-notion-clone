package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/muesli/termenv"

	"notion_ssh_app/internal/app/db"
	"notion_ssh_app/internal/app/models"
	"notion_ssh_app/internal/styles"

	_ "github.com/lib/pq"
)

// Define the main model struct
type Model struct {
	FormModel    *FormModel
	ListView     ListViewModel
	TextareaView TextareaViewModel
	ViewportView ViewportViewModel
	ListItemView models.ListItemViewModel
	CurrentView  int
	Quitting     bool
	LoggedIn     bool
	User         UserDetails
	Dimensions   models.Dimensions
	SplashActive bool
}

type UserDetails struct {
	Username string
	Password string
	user_id  int
}

// Define the form model struct
type FormModel struct {
	Form  *huh.Form
	Style lipgloss.Style
	State huh.FormState
}

// Define the list view model struct
type ListViewModel struct {
	List         list.Model
	ShowSelected bool
}

// Define the textarea view model struct
type TextareaViewModel struct {
	Textarea     textarea.Model
	ShowTextArea bool
}

// Define the viewport view model struct
type ViewportViewModel struct {
	Viewport viewport.Model
	Content  string
}

type SplashFinishedMsg struct{}

// Init method

func (m Model) Init() tea.Cmd {
	if m.SplashActive {
		// If the splash screen is active, return a command to wait for 2 seconds before proceeding
		return tea.Batch(
			tea.Tick(time.Second*5, func(_ time.Time) tea.Msg {
				return SplashFinishedMsg{} // Custom message to signal splash screen end
			}),
		)
	}

	if m.FormModel != nil && m.FormModel.Form != nil {
		// If the form model is not nil, initialize the form
		fmt.Println("about to start the form init")
		return m.FormModel.Form.Init()
	}
	return func() tea.Msg {
		// Fetch the user's list items based on the username stored in the form
		username := m.FormModel.Form.GetString("username")
		// password := m.FormModel.Form.GetString("password")
		fmt.Println(username)
		return db.FetchItems(m.User.user_id)
	}

}

/* VIEW METHODS */
func (m ListViewModel) View() string {
	return styles.ListStyle.Render(m.List.View())
}

// Renders the textarea view
func (m TextareaViewModel) View() string {
	return styles.TextareaStyle.Render(m.Textarea.View())
}

// Renders the viewport view
func (m ViewportViewModel) View() string {

	return styles.ViewportStyle.Render(m.Viewport.View())
}

// // Renders the login form view
// func (m model) View() string {
// 	if m.quitting {
// 		return "exiting the ssh session"
// 	}

// 	// if m.formModel == nil {
// 	// 	return "Starting..."
// 	// }

// 	if m.formModel.state == huh.StateCompleted {
// 		return m.formModel.style.Render("Welcome, " + m.formModel.form.GetString("username") + "!")
// 	}
// 	switch m.currentView {
// 	case 1:
// 		return m.listView.View()
// 	case 2:
// 		return lipgloss.JoinHorizontal(lipgloss.Top, m.textareaView.View(), m.viewportView.View())
// 	case 3:
// 		centeredViewportStyle := lipgloss.NewStyle().
// 			MarginLeft(40).
// 			Render(m.viewportView.View())
// 		return centeredViewportStyle
// 		// return m.viewportView.View()
// 	default:
// 		return m.formModel.form.View()

// 	}

// }

// Renders the login form view
func (m Model) View() string {
	if m.Quitting {
		return "exiting the ssh session"
	}

	formWidth := 50
	formHeight := 10

	if m.SplashActive {
		// Display the splash screen with the service name

		return lipgloss.Place(m.Dimensions.TotalWidth, m.Dimensions.TotalHeight, lipgloss.Center, lipgloss.Center, styles.Logostyle)

	}

	// Prioritize the current view after form submission
	if m.LoggedIn {
		switch m.CurrentView {
		case 1:
			centeredList := lipgloss.Place(m.Dimensions.TotalWidth, m.Dimensions.TotalHeight, lipgloss.Center, lipgloss.Center, m.ListView.View())
			return centeredList
		case 2:
			return lipgloss.JoinHorizontal(lipgloss.Top, m.TextareaView.View(), m.ViewportView.View())
		case 3:
			viewportView := styles.CenteredViewportStyle.Render(m.ViewportView.View())
			centeredViewPort := lipgloss.Place(m.Dimensions.TotalWidth, m.Dimensions.TotalHeight, lipgloss.Center, lipgloss.Center, viewportView)
			return centeredViewPort
		default:
			return m.ListView.View() // Default to list view if logged in
		}
	}

	// Style for the form
	formView := lipgloss.NewStyle().
		Width(formWidth).
		Height(formHeight).
		Border(lipgloss.ThickBorder()).
		Padding(1, 2).
		BorderForeground(lipgloss.Color("#7571F9")). // Purple border
		Foreground(lipgloss.Color("#7571F9")).       // Purple text
		Align(lipgloss.Left)

	// Style for the title
	TitleStyle := lipgloss.NewStyle().
		Width(formWidth).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("#7571F9")) // Purple color

	// Render the banner text stored inside the styles folder
	title := TitleStyle.Render(styles.BannerText)

	// Render the form content
	loginForm := formView.Render(m.FormModel.Form.View())

	// Combine the title and form
	combinedView := lipgloss.JoinVertical(lipgloss.Center, title, loginForm)

	// Center the combined view in the terminal
	finalView := lipgloss.Place(m.Dimensions.TotalWidth, m.Dimensions.TotalHeight, lipgloss.Center, lipgloss.Center, combinedView)

	return finalView
}

/* UPDATE METHODS */
// Update method to handle key presses and window resizing
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update the form if it's not nil
	if m.FormModel != nil {
		f, cmd := m.FormModel.Form.Update(msg)
		m.FormModel.Form = f.(*huh.Form)
		m.FormModel.State = m.FormModel.Form.State
		cmds = append(cmds, cmd)
	}

	// Handle the form state and user login status
	if m.FormModel != nil {
		switch m.FormModel.State {
		case huh.StateAborted:
			return m, tea.Quit

		case huh.StateCompleted:
			if !m.LoggedIn {
				// Get the username and password from the form fields
				username := m.FormModel.Form.GetString("username")
				password := m.FormModel.Form.GetString("password")

				// Attempt to authenticate the user
				userID, err := db.Authenticate(username, password)
				if err != nil {
					// Handle any database errors (e.g., connection issues)
					fmt.Println("Error during authentication:", err)
					// m.ErrorMessage = "An error occurred. Please try again."
				} else if userID != nil {
					// Successfully authenticated; store the user ID and redirect to the list view
					m.User.user_id = *userID
					m.User.Username = username
					m.LoggedIn = true
					m.CurrentView = 1

					// Fetch the user's items
					cmd := func() tea.Msg {
						return db.FetchItems(m.User.user_id)
					}
					cmds = append(cmds, cmd)
				} else {
					// Handle invalid credentials
					fmt.Println("Invalid username or password")
					// m.ErrorMessage = "Invalid username or password"
				}

				// Return the updated model and combined commands
				return m, tea.Batch(cmds...)
			}
		}
	}

	// Handle messages for resizing and input events
	switch msg := msg.(type) {
	case SplashFinishedMsg:
		// Disable splash screen and proceed to form initialization
		m.SplashActive = false
		return m, m.Init() // Re-call Init to initialize the form or fetch items

	// Other cases...

	case tea.WindowSizeMsg:
		// Adjust the sizes of the views based on window size
		m.ListView.List.SetSize(msg.Width-20, msg.Height-10)
		m.ViewportView.Viewport.Width = msg.Width / 2
		m.ViewportView.Viewport.Height = msg.Height - 4
		m.TextareaView.Textarea.SetWidth(msg.Width / 2)
		m.TextareaView.Textarea.SetHeight(msg.Height - 4)
		m.Dimensions.TotalHeight = msg.Height
		m.Dimensions.TotalWidth = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "ctrl+a":
			m.TextareaView.ShowTextArea = !m.TextareaView.ShowTextArea
			if m.TextareaView.ShowTextArea {
				// before opening this view reset the textarea and viewport so user will see fresh empty screens
				m.TextareaView.Textarea.Reset()
				m.ViewportView.Viewport.SetContent("")
				m.CurrentView = 2
			} else {
				m.CurrentView = 1
			}
			return m, nil

		case "ctrl+e":
			if m.TextareaView.ShowTextArea {
				// Get the full content from the textarea
				fullText := m.TextareaView.Textarea.Value()

				// Split the content by lines
				lines := strings.Split(fullText, "\n")

				// Extract the title, description, and content
				var title, desc, content string
				if len(lines) > 0 {
					title = lines[0]
				}
				if len(lines) > 1 {
					desc = lines[1]
				}
				if len(lines) > 2 {
					content = strings.Join(lines[2:], "\n")
				}

				// Create a new item with the extracted values
				newItem := models.ListItemViewModel{
					ItemTitle: title,
					Desc:      desc,
					Content:   content,
				}

				// Add the new item to the database
				err := db.AddItemToDB(newItem)
				if err != nil {
					fmt.Println("Error adding item to database:", err)
				} else {
					fmt.Println("Item added to the database successfully.")
				}

				// Insert the new item into the list and update the view
				m.ListView.List.InsertItem(len(m.ListView.List.Items()), newItem)
				m.TextareaView.ShowTextArea = false
				m.CurrentView = 1
				return m, nil
			}
		case "ctrl+z":
			if m.CurrentView == 1 {
				if i, ok := m.ListView.List.SelectedItem().(models.ListItemViewModel); ok {
					fmt.Println("item number selected is : ", i.ItemTitle)
					fmt.Println("item number selected is : ", i.Description())
					fmt.Println("item number selected is : ", i.Content)
					fmt.Println("item number selected is : ", i.Desc)
					fmt.Println("item number selected is : ", i.Title())

					m.ListItemView = i
					m.CurrentView = 3
					out, _ := glamour.Render(m.ListItemView.Content, "dark") // used glamour to render the markdown in prettier way here
					m.ViewportView.Viewport.SetContent(out)
					// m.ViewportView.Viewport.Style.MarginLeft(30)

				}
				return m, nil
			}
			m.CurrentView = 1
		}

	case tea.MouseMsg:
		if m.CurrentView == 2 {
			var cmd tea.Cmd
			m.ViewportView.Viewport, cmd = m.ViewportView.Viewport.Update(msg)
			return m, cmd
		}

	case models.ItemsMsg:
		var items []list.Item
		for _, i := range msg.Items {
			items = append(items, i)
		}
		m.ListView.List.SetItems(items)
		m.TextareaView.Textarea.Reset()
		m.CurrentView = 1
		return m, nil
	}

	// Update the current view based on the view state
	switch m.CurrentView {
	case 1:
		var cmd tea.Cmd
		m.ListView.List, cmd = m.ListView.List.Update(msg)
		return m, cmd

	case 2:
		var cmd tea.Cmd
		m.TextareaView.Textarea, cmd = m.TextareaView.Textarea.Update(msg)
		out, _ := glamour.Render(m.TextareaView.Textarea.Value(), "dark")
		m.ViewportView.Viewport.SetContent(out)
		return m, cmd

	case 3:
		var cmd tea.Cmd
		m.ViewportView.Viewport, cmd = m.ViewportView.Viewport.Update(msg)
		return m, cmd

	default:
		return m, tea.Batch(cmds...)
	}
}

/* ----------------------------------------------------------------------------------------------------------------------- */

// ListMiddleware returns a Wish middleware that sets up the Bubble Tea program
func ListMiddleware() wish.Middleware {
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}

		form := huh.NewForm(
			huh.NewGroup(

				huh.NewInput().Title("Username").Key("username"),
				huh.NewInput().Title("Password").Key("password").EchoMode(huh.EchoModePassword),
			),
		)

		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 6, 24)
		l.Title = "your notes -> "

		t := textarea.New()
		t.Placeholder = "Title.... \nDescription.....\n"
		t.Focus()
		t.ShowLineNumbers = false
		t.Cursor.Blink = true
		t.CharLimit = 100000

		v := viewport.New(100, 40)
		v.SetContent("Viewport content goes here…")
		m := Model{
			FormModel: &FormModel{
				Form: form,
			},
			SplashActive: true,
			ListView:     ListViewModel{List: l},
			TextareaView: TextareaViewModel{Textarea: t},
			ViewportView: ViewportViewModel{Viewport: v},
		}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
