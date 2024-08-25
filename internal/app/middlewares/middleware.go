package middlewares

import (
	"database/sql"
	"fmt"
	"strings"

	"log"

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

	_ "github.com/lib/pq"
)

// Styles
var DocStyle = lipgloss.NewStyle().Margin(4, 10, 0)

// Define the main model struct
type Model struct {
	FormModel    *FormModel
	ListView     ListViewModel
	TextareaView TextareaViewModel
	ViewportView ViewportViewModel
	ListItemView ListItemViewModel
	CurrentView  int
	Quitting     bool
	LoggedIn     bool
	User         UserDetails
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

// Define the list item view model struct
type ListItemViewModel struct {
	ItemTitle       string
	Desc            string
	Content         string
	ShowItemContent bool
}

// Struct to hold a slice of items
type ItemsMsg struct {
	Items []ListItemViewModel
}

// Methods to fulfill the list.Item interface
func (i ListItemViewModel) FilterValue() string { return i.ItemTitle }
func (i ListItemViewModel) Title() string       { return i.ItemTitle }
func (i ListItemViewModel) Description() string { return i.Desc }

// Init method

func (m Model) Init() tea.Cmd {

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
		return FetchItems(m.User.user_id)
	}

}

/* VIEW METHODS */
func (m ListViewModel) View() string {
	return DocStyle.Render(m.List.View())
}

// Renders the textarea view
func (m TextareaViewModel) View() string {
	textareaStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		// Padding(1, 2).
		MarginTop(2).
		BorderForeground(lipgloss.Color("#d534eb")).
		// Background(lipgloss.Color("#020d14")).
		Foreground(lipgloss.Color("#eb9e34"))

	return textareaStyle.Render(m.Textarea.View())
}

// Renders the viewport view
func (m ViewportViewModel) View() string {
	viewportStyle := lipgloss.NewStyle().
		// Border(lipgloss.ThickBorder(), false, false, true, false).
		//Padding(1, 2).
		MarginTop(2).
		// BorderForeground(lipgloss.Color("63")).
		// Background(lipgloss.Color("#020d14")).
		Foreground(lipgloss.Color("#eb9e34"))

	return viewportStyle.Render(m.Viewport.View())
}

// Renders the individual item view
func (m ListItemViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Margin(4, 10, 0).Height(16).Width(100).Border(lipgloss.NormalBorder(), false, false, true, false).Render(m.Content),
		lipgloss.NewStyle().Height(2).MarginLeft(10).MarginTop(2).Render("ctrl+a: exit alt screen"),
	)
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

	// Prioritize the current view after form submission
	if m.LoggedIn {
		switch m.CurrentView {
		case 1:
			return m.ListView.View()
		case 2:
			return lipgloss.JoinHorizontal(lipgloss.Top, m.TextareaView.View(), m.ViewportView.View())
		case 3:
			centeredViewportStyle := lipgloss.NewStyle().
				MarginLeft(40).
				Render(m.ViewportView.View())
			return centeredViewportStyle
		default:
			return m.ListView.View() // Default to list view if logged in
		}
	}

	// If the form is still active (not submitted), render the form view
	return m.FormModel.Form.View()
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
				userID, err := Authenticate(username, password)
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
						return FetchItems(m.User.user_id)
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
	case tea.WindowSizeMsg:
		// Adjust the sizes of the views based on window size
		m.ListView.List.SetSize(msg.Width-20, msg.Height-10)
		m.ViewportView.Viewport.Width = msg.Width / 2
		m.ViewportView.Viewport.Height = msg.Height - 4
		m.TextareaView.Textarea.SetWidth(msg.Width / 2)
		m.TextareaView.Textarea.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "ctrl+a":
			m.TextareaView.ShowTextArea = !m.TextareaView.ShowTextArea
			if m.TextareaView.ShowTextArea {
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
				newItem := ListItemViewModel{
					ItemTitle: title,
					Desc:      desc,
					Content:   content,
				}

				// Add the new item to the database
				err := AddItemToDB(newItem)
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
				if i, ok := m.ListView.List.SelectedItem().(ListItemViewModel); ok {
					fmt.Println("item number selected is : ", i.ItemTitle)
					fmt.Println("item number selected is : ", i.Description())
					fmt.Println("item number selected is : ", i.Content)
					fmt.Println("item number selected is : ", i.Desc)
					fmt.Println("item number selected is : ", i.Title())

					m.ListItemView = i
					m.CurrentView = 3
					m.ViewportView.Viewport.SetContent(i.Content)

					m.ViewportView.Viewport.Style.MarginLeft(20)

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

	case ItemsMsg:
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

		style := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(1, 2).
			BorderForeground(lipgloss.Color("#444444")).
			Foreground(lipgloss.Color("#7571F9"))

		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 6, 24)
		l.Title = "your notes -> "
		t := textarea.New()
		t.Placeholder = "Enter some text…"
		t.Focus()
		t.ShowLineNumbers = true
		t.Cursor.Blink = true
		t.CharLimit = 10000
		v := viewport.New(100, 40)
		v.SetContent("Viewport content goes here…")
		m := Model{
			FormModel: &FormModel{
				Form:  form,
				Style: style,
			},
			ListView:     ListViewModel{List: l},
			TextareaView: TextareaViewModel{Textarea: t},
			ViewportView: ViewportViewModel{Viewport: v},
		}

		return tea.NewProgram(m, tea.WithInput(s), tea.WithOutput(s), tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

// OpenDB opens and returns a database connection.
func OpenDB() (*sql.DB, error) {
	connStr := "postgresql://article%20list_owner:UnHc9jlDV7Oo@ep-orange-bush-a19fqe45.ap-southeast-1.aws.neon.tech/article%20list?sslmode=require"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AddItemToDB adds a new item to the database.
func AddItemToDB(item ListItemViewModel) error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO items (ItemTitle, description, content, user_id) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(query, item.ItemTitle, item.Desc, item.Content, 1)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate checks if the user exists and returns the user ID if valid, otherwise returns nil
func Authenticate(username, password string) (*int, error) {
	db, err := OpenDB()
	if err != nil {
		return nil, err // Return nil for the ID and the error
	}
	defer db.Close()

	var userID int
	query := `
        SELECT id FROM users WHERE username = $1 AND password = $2 LIMIT 1;
    `
	err = db.QueryRow(query, username, password).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err // Some other error occurred
	}

	return &userID, nil // User found, return their ID
}

// FetchItems fetches the items for a specific user from the database
func FetchItems(userID int) tea.Msg {
	db, err := OpenDB() // OpenDB is a function that connects to the database
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return ItemsMsg{Items: []ListItemViewModel{}}
	}
	defer db.Close()

	// Prepare the query to fetch items for the given userID
	query := `
        SELECT ItemTitle, description, content 
        FROM items 
        WHERE user_id = $1;
    `
	rows, err := db.Query(query, userID)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		return ItemsMsg{Items: []ListItemViewModel{}}
	}
	defer rows.Close()

	// Iterate over the rows and create a list of ListItemViewModel
	var userItems []ListItemViewModel
	for rows.Next() {
		var item ListItemViewModel
		if err := rows.Scan(&item.ItemTitle, &item.Desc, &item.Content); err != nil {
			fmt.Println("Error scanning row:", err)
			return ItemsMsg{Items: []ListItemViewModel{}}
		}
		userItems = append(userItems, item)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating over rows:", err)
		return ItemsMsg{Items: []ListItemViewModel{}}
	}

	// Return the fetched items in an ItemsMsg
	return ItemsMsg{Items: userItems}
}

// Example usage to check the database connection
func CheckDBVersion() {
	db, err := OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT version()")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var version string
	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("version=%s\n", version)
}
