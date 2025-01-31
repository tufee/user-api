package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"userapi/db"
	"userapi/models"
)

func create(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error to decode JSON", http.StatusBadRequest)
		return
	}

	con, err := db.GetConnection()

	if err != nil {
		http.Error(w, "Failed to get db connection", http.StatusInternalServerError)
	}
	defer con.Close()

	id, err := db.CreateUser(con, user.Name, user.Email)
	if err != nil {
		http.Error(w, "Error to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "User created with ID: %d\n", id)
}

func findAll(w http.ResponseWriter, r *http.Request) {
	con, err := db.GetConnection()

	if err != nil {
		http.Error(w, "Failed to get db connection", http.StatusInternalServerError)
		return
	}
	defer con.Close()

	users, err := db.GetUsers(con)
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode users to json", http.StatusInternalServerError)
		return
	}

}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	con, err := db.GetConnection()
	if err != nil {
		http.Error(w, "Failed to get db connection", http.StatusInternalServerError)
		return
	}
	defer con.Close()

	path := strings.TrimPrefix(r.URL.Path, "/delete/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	result, err := db.DeleteUser(con, id)
	if err != nil {
		http.Error(w, "Error to delete user", http.StatusInternalServerError)
		return
	}

	RowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error to verify affected registers", http.StatusInternalServerError)
		return
	}

	if RowsAffected == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.ID <= 0 {
		http.Error(w, "ID must be greater than 0", http.StatusBadRequest)
		return
	}

	if user.Name == "" && user.Email == "" {
		http.Error(w, "At least one field must be provided", http.StatusBadRequest)
		return
	}

	query, args, err := buildUpdateQuery(user.ID, user.Name, user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	con, err := db.GetConnection()

	if err != nil {
		http.Error(w, "Failed to get db connection", http.StatusInternalServerError)
	}
	defer con.Close()

	err = db.UpdateUser(con, query, args)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User updated successfully!"))
}

func buildUpdateQuery(id int, name, email string) (string, []interface{}, error) {
	query := "UPDATE users SET"
	args := []interface{}{}

	if name != "" {
		query += " name = ?,"
		args = append(args, name)
	}
	if email != "" {
		query += " email = ?,"
		args = append(args, email)
	}

	query = strings.TrimSuffix(query, ",")

	query += " WHERE id = ?"
	args = append(args, id)

	return query, args, nil
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("POST /create", create)
	router.HandleFunc("GET /find", findAll)
	router.HandleFunc("DELETE /delete/{id}", deleteUser)
	router.HandleFunc("PUT /update", updateUser)

	server := http.Server{
		Addr:    ":3000",
		Handler: router,
	}

	fmt.Println("Server listening on port :3000")
	server.ListenAndServe()
}
