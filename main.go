package main

import (
	"file-handler/processor"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	uploadDir  = "D:/file-handler/uploads"
	outputsDir = "D:/file-handler/outputs"
)

func init() {
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(outputsDir, 0755)
	wd, _ := os.Getwd()
	log.Printf("Working directory: %s", wd)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No 'file' field in form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	log.Printf("Detected content type: %s", contentType)

	// name generation
	baseName := time.Now().Format("02-01-2006-15-04-05") + "-" + sanitizeFilename(strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)))

	inputPath := filepath.Join(uploadDir, baseName+filepath.Ext(header.Filename))
	outputPath := filepath.Join(outputsDir, baseName+".json")

	// saving primary file
	err = os.WriteFile(inputPath, fileBytes, 0644)
	if err != nil {
		log.Printf("Failed to save input file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	var jsonData []byte

	switch {
	case strings.HasPrefix(contentType, "text/csv") || strings.HasSuffix(strings.ToLower(header.Filename), ".csv"):
		jsonData, err = processor.CSVToJSONBytes(inputPath)

	case contentType == "image/png":
		jsonData, err = processor.ProcessPNG(inputPath)

	case strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") || strings.HasSuffix(strings.ToLower(header.Filename), ".xls"):
		jsonData, err = processor.ProcessExcel(inputPath)

	default:
		http.Error(w, "Unsupported file type: "+header.Filename+" ("+contentType+")", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("Processing failed: %v", err)
		http.Error(w, "Processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// saving json file to outputDir
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		log.Printf("Failed to save output JSON: %v", err)
		// Не прерываем ответ — клиент всё равно получит данные
	}

	// json sending
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	if name == "" {
		return "unknown"
	}
	return name
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
