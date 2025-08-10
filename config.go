package main

import (
	"os"
	"strings"
)

type Config struct {
	Port            string
	AllowedExts     []string
	MaxFilesPerTask int
	MaxActiveTasks  int
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	exts := os.Getenv("ALLOWED_EXTS")
	if exts == "" {
		exts = ".pdf,.jpeg,.jpg"
	}
	maxFiles := 3
	maxTasks := 3
	return &Config{
		Port:            port,
		AllowedExts:     strings.Split(exts, ","),
		MaxFilesPerTask: maxFiles,
		MaxActiveTasks:  maxTasks,
	}
}
