package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const (
	memoryDir = ".nanobot/workspace/memory"
	skillsDir = ".nanobot/workspace/skills"
)

type FileContent struct {
	Content string `json:"content"`
}

func getMemoryFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("file")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Failed to get user home directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(homeDir, memoryDir, fileName)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(FileContent{Content: string(content)})
}

func updateMemoryFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.PathValue("file")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Failed to get user home directory", http.StatusInternalServerError)
		return
	}

	var fileContent FileContent
	if err := json.NewDecoder(r.Body).Decode(&fileContent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(homeDir, memoryDir, fileName)
	if err := ioutil.WriteFile(filePath, []byte(fileContent.Content), 0644); err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
