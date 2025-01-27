package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CUHK-SE-Group/rcabench/config"
)

type ExculdeRule struct {
	Pattern string
	IsGlob  bool
}

// 添加文件到 ZIP
func AddToZip(zipWriter *zip.Writer, srcPath string, zipPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	entry, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(entry, file)
	return err
}

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

// 安全检查防止路径遍历攻击
func IsAllowedPath(path string) bool {
	allowedRoot := config.GetString("nfs.path")
	rel, err := filepath.Rel(allowedRoot, path)
	return err == nil && !strings.Contains(rel, "..")
}

// 判断文件是否匹配排除规则
func MatchFile(fileName string, rule ExculdeRule) bool {
	if rule.IsGlob {
		match, _ := filepath.Match(rule.Pattern, fileName)
		return match
	}
	return fileName == rule.Pattern
}
