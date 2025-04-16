package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
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
func AddToZip(zipWriter *zip.Writer, fileInfo fs.FileInfo, srcPath string, zipPath string) error {
	fileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	fileHeader.Name = zipPath
	fileHeader.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(fileHeader)
	if err != nil {
		return err
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
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

// extractZip extracts a zip file to the specified destination directory
func ExtractZip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Check for path traversal vulnerabilities
		filePath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// extractTarGz extracts a tar.gz file to the specified destination directory
func ExtractTarGz(tarGzFile, destDir string) error {
	file, err := os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Check for path traversal vulnerabilities
		filePath := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			outFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
