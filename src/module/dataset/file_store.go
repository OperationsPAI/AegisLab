package dataset

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	"aegis/model"
	"aegis/utils"
)

type DatapackFileStore struct {
	basePath string
}

func NewDatapackFileStore() *DatapackFileStore {
	return &DatapackFileStore{basePath: config.GetString("jfs.dataset_path")}
}

func (s *DatapackFileStore) PackageToZip(zipWriter *zip.Writer, datapacks []model.FaultInjection, excludeRules []utils.ExculdeRule) error {
	for i := range datapacks {
		if err := s.packageDatapackToZip(zipWriter, &datapacks[i], excludeRules); err != nil {
			return err
		}
	}
	return nil
}

func (s *DatapackFileStore) packageDatapackToZip(zipWriter *zip.Writer, datapack *model.FaultInjection, excludeRules []utils.ExculdeRule) error {
	if datapack.State < consts.DatapackBuildSuccess {
		return fmt.Errorf("datapack %s is not in a downloadable state", datapack.Name)
	}

	workDir := filepath.Join(s.basePath, datapack.Name)
	if !utils.IsAllowedPath(workDir) {
		return fmt.Errorf("invalid path access to %s", workDir)
	}

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
		return fmt.Errorf("failed to package datapack %s: %w", datapack.Name, err)
	}

	return nil
}
