package utils

import "os"

func GetSubFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
