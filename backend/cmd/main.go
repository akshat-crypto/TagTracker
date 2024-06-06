package main

import (
	"log"
	"net/http"

	"github.com/akshat-crypto/TagTracker/backend/internal/db"
	"github.com/akshat-crypto/TagTracker/backend/internal/handlers"
	"github.com/gorilla/mux"
)

func main() {

	//Initialize the database
	db.InitDB()

	r := mux.NewRouter()
	r.HandleFunc("/scan", handlers.ScanHandler).Methods("POST")
	r.HandleFunc("/resources", handlers.ResourcesHandler).Methods("GET")
	r.HandleFunc("/update-tags", handlers.UpdateTagsHandler).Methods("POST")

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server Run Failed:", err)
	}
}
