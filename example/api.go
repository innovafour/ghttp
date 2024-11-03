package main

import (
	"encoding/json"
	"net/http"
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error getting form", http.StatusInternalServerError)
		return
	}

	name := r.FormValue("name")
	age := r.FormValue("age")

	w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusInternalServerError)

	err := json.NewEncoder(w).Encode(map[string]string{"name": name, "age": age})
	if err != nil {
		http.Error(w, "Error encoding json", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]string{"name": name, "age": age})
}
