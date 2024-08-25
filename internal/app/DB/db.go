package db

import (
	"database/sql"
	"fmt"

	"log"

	tea "github.com/charmbracelet/bubbletea"

	"notion_ssh_app/internal/app/models"

	_ "github.com/lib/pq"
)

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
func AddItemToDB(item models.ListItemViewModel) error {
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
		return models.ItemsMsg{Items: []models.ListItemViewModel{}}
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
		return models.ItemsMsg{Items: []models.ListItemViewModel{}}
	}
	defer rows.Close()

	// Iterate over the rows and create a list of models.ListItemViewModel
	var userItems []models.ListItemViewModel
	for rows.Next() {
		var item models.ListItemViewModel
		if err := rows.Scan(&item.ItemTitle, &item.Desc, &item.Content); err != nil {
			fmt.Println("Error scanning row:", err)
			return models.ItemsMsg{Items: []models.ListItemViewModel{}}
		}
		userItems = append(userItems, item)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating over rows:", err)
		return models.ItemsMsg{Items: []models.ListItemViewModel{}}
	}

	// Return the fetched items in an ItemsMsg
	return models.ItemsMsg{Items: userItems}
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
