package main

import (
	"log"
	"net/http"
	"strings"
)

func main() {
	config := LoadConfig()
	tm := NewTaskManager(config)

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			createTaskHandler(tm)(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/files") && r.Method == http.MethodPost:
			addFileHandler(tm)(w, r)
		case strings.HasSuffix(path, "/archive") && r.Method == http.MethodGet:
			downloadArchiveHandler(tm)(w, r)
		case r.Method == http.MethodGet:
			getTaskStatusHandler(tm)(w, r)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	log.Printf("Server started at :%s\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
