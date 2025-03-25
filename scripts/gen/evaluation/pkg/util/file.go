package util

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadCSVFile(filePath string) (*[][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	return &records, nil
}

func IsJSONByExt(filePath string) bool {
	ext := filepath.Ext(filePath)
	return strings.ToLower(ext) == ".json"
}

func OutputToJSON(results any, filePath string) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON serialization failed: %v", err)
	}

	if dir := filepath.Dir(filePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("directory creation failed: %v", err)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("file creation failed: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("file closed err: %v", closeErr)
		}
	}()

	if _, err = file.Write(jsonData); err != nil {
		log.Fatalf("failed to write file: %v", err)
	}

	log.Printf("results successfully saved to: %s", filePath)

	return nil
}
