package containermodule

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestHelmFileStoreSaveChartAndValueFile(t *testing.T) {
	tmpDir := t.TempDir()
	viper.Set("jfs.dataset_path", tmpDir)

	store := &HelmFileStore{basePath: tmpDir}
	fileHeader := newMultipartFileHeader(t, "chart.tgz", []byte("chart-bytes"))

	chartPath, checksum, err := store.SaveChart("pedestal", fileHeader)
	if err != nil {
		t.Fatalf("SaveChart failed: %v", err)
	}
	if checksum == "" {
		t.Fatalf("expected checksum to be populated")
	}
	if !filepath.IsAbs(chartPath) && filepath.Dir(chartPath) == "." {
		t.Fatalf("expected chart path to include target directory, got %s", chartPath)
	}

	chartContent, err := os.ReadFile(chartPath)
	if err != nil {
		t.Fatalf("read saved chart: %v", err)
	}
	if string(chartContent) != "chart-bytes" {
		t.Fatalf("unexpected chart content: %s", string(chartContent))
	}

	valueHeader := newMultipartFileHeader(t, "values.yaml", []byte("key: value\n"))
	valuePath, err := store.SaveValueFile("pedestal", valueHeader, "")
	if err != nil {
		t.Fatalf("SaveValueFile failed: %v", err)
	}

	valueContent, err := os.ReadFile(valuePath)
	if err != nil {
		t.Fatalf("read saved values file: %v", err)
	}
	if string(valueContent) != "key: value\n" {
		t.Fatalf("unexpected values content: %s", string(valueContent))
	}
}

func newMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(body.Len()) + 1024); err != nil {
		t.Fatalf("parse multipart form: %v", err)
	}

	fileHeaders := req.MultipartForm.File["file"]
	if len(fileHeaders) != 1 {
		t.Fatalf("expected one file header, got %d", len(fileHeaders))
	}
	return fileHeaders[0]
}
