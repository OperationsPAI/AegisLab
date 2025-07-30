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

	"github.com/LGU-SE-Internal/rcabench/config"
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

func GetAllSubDirectories(root string) ([]string, error) {
	var directories []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// This is a directory
			path := filepath.Join(root, entry.Name())
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			directories = append(directories, absPath)
		}
	}

	return directories, nil
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

func CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("failed to get directory info: %v", err)
			}
			return os.MkdirAll(dstPath, info.Mode())
		} else {
			return CopyFile(path, dstPath)
		}
	})
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

// ExtractZip 解压zip文件，如果只有一个顶级目录则提升其内容到根目录
func ExtractZip(zipFile, destDir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	var topLevelDir string
	allInSingleDir := true

	for _, f := range r.File {
		parts := strings.Split(f.Name, "/")
		if len(parts) == 1 && !f.FileInfo().IsDir() {
			allInSingleDir = false
			break
		}
		if topLevelDir == "" {
			topLevelDir = parts[0]
		} else if topLevelDir != parts[0] {
			allInSingleDir = false
			break
		}
	}

	for _, f := range r.File {
		var filePath string

		if allInSingleDir && topLevelDir != "" {
			relativePath := strings.TrimPrefix(f.Name, topLevelDir+"/")
			if relativePath == "" {
				continue
			}
			filePath = filepath.Join(destDir, relativePath)
		} else {
			filePath = filepath.Join(destDir, f.Name)
		}

		// 安全检查
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

// ExtractTarGz 解压tar.gz文件，如果只有一个顶级目录则提升其内容到根目录
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

	var headers []*tar.Header
	var topLevelDir string
	allInSingleDir := true

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		headers = append(headers, header)

		parts := strings.Split(header.Name, "/")
		if len(parts) == 1 && header.Typeflag == tar.TypeReg {
			allInSingleDir = false
			break
		}
		if topLevelDir == "" {
			topLevelDir = parts[0]
		} else if topLevelDir != parts[0] {
			allInSingleDir = false
			break
		}
	}

	file.Close()
	file, err = os.Open(tarGzFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err = gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr = tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var filePath string

		if allInSingleDir && topLevelDir != "" {
			relativePath := strings.TrimPrefix(header.Name, topLevelDir+"/")
			if relativePath == "" {
				continue // 跳过顶级目录本身
			}

			filePath = filepath.Join(destDir, relativePath)
		} else {
			filePath = filepath.Join(destDir, header.Name)
		}

		// 安全检查
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
