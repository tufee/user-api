package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"userapi/models"
)

func GetConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./user.db")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	createTableSql := `
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE
	);`

	_, err = db.Exec(createTableSql)

	if err != nil {
		log.Fatalf("Error to create table: %v", err)
	}

	fmt.Println("Table created or already exists")

	return db, nil
}

func CreateUser(db *sql.DB, name, email string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO users(name, email) VALUES(?, ?)")
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	result, err := stmt.Exec(name, email)
	if err != nil {
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}

	return id, nil

}

func GetUsers(db *sql.DB) ([]models.User, error) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, fmt.Errorf("Error to find users: %w", err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("Error to scan users: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error after interaction: %w", err)
	}

	return users, nil
}

func DeleteUser(db *sql.DB, id int) (sql.Result, error) {
	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("Error to delete user")
	}

	return result, err
}

func UpdateUser(db *sql.DB, query string, args []interface{}) error {
	_, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("Database error: %v", err)
	}
	return nil
}
