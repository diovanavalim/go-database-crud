package main

import (
	"databases/server"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	port := ":8686"

	router := mux.NewRouter()

	router.HandleFunc("/user", server.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/user", server.GetUsers).Methods(http.MethodGet)
	router.HandleFunc("/user/{id}", server.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/user/{id}", server.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/user/{id}", server.DeleteUser).Methods(http.MethodDelete)

	fmt.Printf("Server connected! Listenning on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
