package injection

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"aegis/utils"

	"github.com/spf13/viper"
)

func TestDatapackStoreBuildTreeAndOpenFile(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("jfs.dataset_path", tmpDir)
	store := &DatapackStore{basePath: tmpDir}

	root := filepath.Join(tmpDir, "dp-one")
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "data.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	resp, err := store.BuildFileTree("dp-one", "", 12)
	if err != nil {
		t.Fatalf("BuildFileTree failed: %v", err)
	}
	if resp.FileCount != 1 || resp.DirCount != 1 {
		t.Fatalf("unexpected counts: files=%d dirs=%d", resp.FileCount, resp.DirCount)
	}

	name, contentType, size, reader, err := store.OpenFile("dp-one", "nested/data.txt")
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer func() { _ = reader.Close() }()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if name != "data.txt" || contentType != "text/plain" || size != int64(len("hello")) || string(content) != "hello" {
		t.Fatalf("unexpected file result: %s %s %d %q", name, contentType, size, string(content))
	}
}

func TestDatapackStorePackageUsesExcludeRules(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("jfs.dataset_path", tmpDir)
	store := &DatapackStore{basePath: tmpDir}

	root := filepath.Join(tmpDir, "dp-two")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write keep: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "drop.log"), []byte("drop"), 0o644); err != nil {
		t.Fatalf("write drop: %v", err)
	}

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	if err := store.Package(zw, "dp-two", []utils.ExculdeRule{{Pattern: "*.log", IsGlob: true}}); err != nil {
		t.Fatalf("Package failed: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	if len(zr.File) != 1 || filepath.Base(zr.File[0].Name) != "keep.txt" {
		t.Fatalf("unexpected zip entries: %+v", zr.File)
	}
}
