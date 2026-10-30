package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type API struct {
	store *Store
}

func (a *API) getPerson(w http.ResponseWriter, r *http.Request) {
	// 1. Read {id} from the URL path and convert it to an int
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest) // 400
		return
	}

	// 2. Look up the person
	person, err := a.store.findByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound) // 404
		return
	}

	// 3. Send the person back as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

func (a *API) getChildren(w http.ResponseWriter, r *http.Request) {
	// 1. Read and convert the id: same as getPerson
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest) // 400
		return
	}

	// 2. Call the store
	children, err := a.store.Children(id)
	if err != nil {
	// difference #1: which status code?
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	// difference #2: the nil-slice problem (see below)
	if children == nil {
		children = []Person{}
	}

	// 3. Send JSON: same as getPerson, but encode `children`
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(children)
}

func (a *API) getSiblings(w http.ResponseWriter, r *http.Request) {
	// 1. Read and convert the id: same as getPerson
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest) // 400
		return
	}

	// 2. Call the store
	siblings, err := a.store.Siblings(id)
	if err != nil {
	// difference #1: which status code?
		http.Error(w, err.Error(), http.StatusNotFound) // 404
		return
	}

	// difference #2: the nil-slice problem (see below)
	if siblings == nil {
		siblings = []Person{}
	}

	// 3. Send JSON: same as getPerson, but encode `siblings`
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(siblings)
}