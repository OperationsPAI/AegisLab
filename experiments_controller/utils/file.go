package utils

import (
	"os"
	"path/filepath"
)

// 获取所有子目录
func GetAllSubDirectories(root string) ([]string, error) {
	var directories []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != root {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			directories = append(directories, absPath)
		}

		return nil
	})

	return directories, err
}
