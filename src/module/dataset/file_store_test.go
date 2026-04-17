package datasetmodule

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"aegis/consts"
	"aegis/model"
	"aegis/utils"

	"github.com/spf13/viper"
)

func TestDatapackFileStorePackageToZip(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("jfs.dataset_path", tmpDir)

	datapackDir := filepath.Join(tmpDir, "datapack-a")
	if err := os.MkdirAll(filepath.Join(datapackDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir datapack dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(datapackDir, "nested", "keep.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write datapack file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(datapackDir, "skip.log"), []byte("skip"), 0o644); err != nil {
		t.Fatalf("write excluded file: %v", err)
	}

	store := &DatapackFileStore{basePath: tmpDir}
	buf := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buf)
	err := store.PackageToZip(zipWriter, []model.FaultInjection{{
		Name:  "datapack-a",
		State: consts.DatapackBuildSuccess,
	}}, []utils.ExculdeRule{{Pattern: "*.log", IsGlob: true}})
	if err != nil {
		t.Fatalf("PackageToZip failed: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("open zip reader: %v", err)
	}

	files := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open zip file %s: %v", file.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip file %s: %v", file.Name, err)
		}
		files[file.Name] = string(content)
	}

	expected := filepath.ToSlash(filepath.Join(consts.DownloadFilename, "datapack-a", "nested", "keep.txt"))
	if files[expected] != "hello" {
		t.Fatalf("expected zip to contain %s with hello, got %q", expected, files[expected])
	}

	excluded := filepath.ToSlash(filepath.Join(consts.DownloadFilename, "datapack-a", "skip.log"))
	if _, ok := files[excluded]; ok {
		t.Fatalf("expected %s to be excluded", excluded)
	}
}
