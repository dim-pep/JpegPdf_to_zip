package main

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func createTaskHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := tm.CreateTask()
		if err != nil {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": task.ID})
	}
}

func addFileHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		id = strings.TrimSuffix(id, "/files")
		var req struct {
			URL string `json:"url"`
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)
		if req.URL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}
		err := tm.AddFileToTask(id, req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getTaskStatusHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		task, ok := tm.GetTask(id)
		if !ok {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(task)
	}
}

func downloadArchiveHandler(tm *TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		id = strings.TrimSuffix(id, "/archive")
		task, ok := tm.GetTask(id)
		if !ok || task.Status != "done" {
			http.Error(w, "archive not ready", http.StatusNotFound)
			return
		}
		archivePath := filepath.Join("archives", id+".zip")
		http.ServeFile(w, r, archivePath)
	}
}
