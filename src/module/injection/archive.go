package injection

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"path/filepath"

	"aegis/consts"
	"aegis/utils"
)

func packageDatapackDirectoryToZip(zipWriter *zip.Writer, workDir string, excludeRules []utils.ExculdeRule) error {
	err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
		if err != nil || dir.IsDir() {
			return err
		}

		relPath, _ := filepath.Rel(workDir, path)
		fullRelPath := filepath.Join(consts.DownloadFilename, filepath.Base(workDir), relPath)
		fileName := filepath.Base(path)

		for _, rule := range excludeRules {
			if utils.MatchFile(fileName, rule) {
				return nil
			}
		}

		fileInfo, err := dir.Info()
		if err != nil {
			return err
		}

		return utils.AddToZip(zipWriter, fileInfo, path, filepath.ToSlash(fullRelPath))
	})
	if err != nil {
		return fmt.Errorf("failed to package datapack directory %s: %w", filepath.Base(workDir), err)
	}

	return nil
}
