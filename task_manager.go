package main

import (
	"archive/zip"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string            `json:"id"`
	Files     []string          `json:"files"`
	Status    string            `json:"status"`
	Errors    map[string]string `json:"errors"`
	Archive   string            `json:"archive_url,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	Completed bool              `json:"-"`
}

type TaskManager struct {
	Tasks       map[string]*Task
	Mutex       sync.Mutex
	ActiveTasks int
	Config      *Config
}

func GenerateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func NewTaskManager(cfg *Config) *TaskManager {
	return &TaskManager{
		Tasks:  make(map[string]*Task),
		Config: cfg,
	}
}

func (tm *TaskManager) CreateTask() (*Task, error) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	if tm.ActiveTasks >= tm.Config.MaxActiveTasks {
		return nil, errors.New("server is busy, try later")
	}
	id := GenerateID()
	task := &Task{
		ID:        id,
		Files:     []string{},
		Status:    "pending",
		Errors:    make(map[string]string),
		CreatedAt: time.Now(),
	}
	tm.Tasks[id] = task
	tm.ActiveTasks++
	return task, nil
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	task, ok := tm.Tasks[id]
	return task, ok
}

func (tm *TaskManager) AddFileToTask(id, url string) error {
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	task, ok := tm.Tasks[id]
	if !ok {
		return errors.New("task not found")
	}
	if len(task.Files) >= tm.Config.MaxFilesPerTask {
		return errors.New("file limit reached")
	}
	task.Files = append(task.Files, url)
	if len(task.Files) == tm.Config.MaxFilesPerTask && task.Status == "pending" {
		task.Status = "in_progress"
		go tm.processTask(task)
	}
	return nil
}

func (tm *TaskManager) processTask(task *Task) {
	_, errorsMap := DownloadAndArchive(task.ID, task.Files, tm.Config.AllowedExts)
	tm.Mutex.Lock()
	defer tm.Mutex.Unlock()
	task.Errors = errorsMap
	if len(errorsMap) == len(task.Files) {
		task.Status = "error"
	} else {
		task.Status = "done"
		task.Archive = "http://localhost:8080/tasks/" + task.ID + "/archive"
	}
	task.Completed = true
	tm.ActiveTasks--
}

func DownloadAndArchive(taskID string, urls []string, allowedExts []string) (string, map[string]string) {
	tmpDir := filepath.Join("tmp", taskID)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	filesToArchive := []string{}
	errorsMap := make(map[string]string)

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			errorsMap[url] = "download error"
			continue
		}
		defer resp.Body.Close()

		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]
		if !IsAllowedExt(filename, allowedExts) {
			errorsMap[url] = "file type not allowed"
			continue
		}
		filePath := filepath.Join(tmpDir, filename)
		out, err := os.Create(filePath)
		if err != nil {
			errorsMap[url] = "file create error"
			continue
		}
		_, err = io.Copy(out, resp.Body)
		out.Close()
		if err != nil {
			errorsMap[url] = "file save error"
			continue
		}
		filesToArchive = append(filesToArchive, filePath)
	}

	archivePath := filepath.Join("archives", taskID+".zip")
	os.MkdirAll("archives", 0755)
	zipFile, err := os.Create(archivePath)
	if err != nil {
		return "", errorsMap
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	for _, file := range filesToArchive {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		defer f.Close()
		w, err := zipWriter.Create(filepath.Base(file))
		if err != nil {
			continue
		}
		io.Copy(w, f)
	}
	zipWriter.Close()
	return archivePath, errorsMap
}

func IsAllowedExt(filename string, allowedExts []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, a := range allowedExts {
		if ext == a {
			return true
		}
	}
	return false
}
