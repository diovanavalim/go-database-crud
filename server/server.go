package server

import (
	"databases/database"
	"databases/dto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request body read failed"))
		return
	}

	var user dto.User

	if err = json.Unmarshal(requestBody, &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not unmarshall request body"))
		return
	}

	database, err := database.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not connect on application's database"))
		return
	}
	defer database.Close()

	stmt, err := database.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not prepare query statement"))
		return
	}
	defer stmt.Close()

	insert, err := stmt.Exec(user.Name, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not execute insert query"))
		return
	}

	id, err := insert.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not find last insert ID"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("User successfully inserted! ID: %d", id)))
	return
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	database, err := database.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not connect on application's database"))
		return
	}

	defer database.Close()

	rows, err := database.Query("SELECT * FROM users")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not execute select query"))
		return
	}

	defer rows.Close()

	var users []dto.User

	for rows.Next() {
		var user dto.User

		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not scan one or more users"))
			return
		}

		users = append(users, user)
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not encode JSON"))
		return
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	ID, err := strconv.ParseInt(params["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse ID param"))
		return
	}

	db, err := database.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not connect to database"))
		return
	}

	row, err := db.Query("SELECT * FROM users WHERE id = ?", ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not execute select query"))
		return
	}

	var user dto.User
	if row.Next() {
		if err := row.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not scan user"))
			return
		}
	}

	if user.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Could not find user for ID %d", ID)))
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not encode JSON"))
		return
	}
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	ID, err := strconv.ParseInt(params["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse ID param"))
		return
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not read request body"))
		return
	}

	var user dto.User
	if err := json.Unmarshal(requestBody, &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not unmarshal JSON body"))
		return
	}

	db, err := database.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not connect to application's database"))
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE users SET name = ?, email = ? WHERE id = ?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not prepare query statement"))
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(user.Name, user.Email, ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not update user"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	ID, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not parse ID param"))
		return
	}

	db, err := database.Connect()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not connect on application's database"))
		return
	}

	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not prepare statement"))
		return
	}

	defer stmt.Close()

	if _, err := stmt.Exec(ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Could not delete user of ID %d", ID)))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
