package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "github.com/jackc/pgx/v5/stdlib"
)


func main() {
	// 1. Connect to the database
	db, err := sql.Open("pgx", "postgres://Mac@localhost:5432/family_tree")
	if err != nil {
		log.Fatal("open:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("ping:", err)
	}

	store := NewStore(db)

	api := &API{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people/{id}", api.getPerson)
	mux.HandleFunc("GET /people/{id}/children", api.getChildren)
	mux.HandleFunc("GET /people/{id}/siblings", api.getSiblings)

	fmt.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}



