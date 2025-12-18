package zip2jsons_test

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	bj "github.com/takanoriyanagitani/go-blob2json"
	"github.com/takanoriyanagitani/go-zip2blobs2jsons"
)

func TestZipItem_ToBlob(t *testing.T) {
	t.Parallel()

	t.Run("basic conversion", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		header := &zip.FileHeader{
			Name:     "item.txt",
			Method:   zip.Deflate,
			Modified: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		f, err := w.CreateHeader(header)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write([]byte("item content"))
		if err != nil {
			t.Fatalf("Failed to write to file in zip: %v", err)
		}
		err = w.Close()
		if err != nil {
			t.Fatalf("Failed to close zip writer: %v", err)
		}

		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		if len(r.File) == 0 {
			t.Fatalf("No files found in zip reader.")
		}

		var item zip2jsons.ZipItem = zip2jsons.ZipItem{File: r.File[0]}
		var bldr bj.BlobBuilder = bj.BlobBuilder{
			ContentType: "text/plain",
			MaxBytes:    1024,
		}

		blob, err := item.ToBlob(bldr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if blob == nil {
			t.Fatalf("Expected blob, got nil")
		}

		if blob.Name != "item.txt" {
			t.Errorf("Expected name 'item.txt', got '%s'", blob.Name)
		}
		if blob.Body != "aXRlbSBjb250ZW50" { // base64 of "item content"
			t.Errorf("Expected body 'aXRlbSBjb250ZW50', got '%s'", blob.Body)
		}
		if blob.ContentType != "text/plain" {
			t.Errorf("Expected content type 'text/plain', got '%s'", blob.ContentType)
		}
		if *blob.ContentLength != 12 {
			t.Errorf("Expected content length 12, got %d", *blob.ContentLength)
		}
		expectedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		if !blob.LastModified.Equal(expectedTime) {
			t.Errorf("Expected last modified %v, got %v", expectedTime, blob.LastModified)
		}
	})

	t.Run("file content exceeds MaxBytes", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		header := &zip.FileHeader{
			Name:     "large_item.txt",
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		f, err := w.CreateHeader(header)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write([]byte(strings.Repeat("a", 100))) // 100 bytes
		if err != nil {
			t.Fatalf("Failed to write to file in zip: %v", err)
		}
		err = w.Close()
		if err != nil {
			t.Fatalf("Failed to close zip writer: %v", err)
		}

		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		var item zip2jsons.ZipItem = zip2jsons.ZipItem{File: r.File[0]}
		var bldr bj.BlobBuilder = bj.BlobBuilder{
			ContentType: "text/plain",
			MaxBytes:    50, // Limit to 50 bytes
		}

		blob, err := item.ToBlob(bldr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if blob == nil {
			t.Fatalf("Expected blob, got nil")
		}

		// Check that the content is truncated to MaxBytes
		expectedBody := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 50)))
		if blob.Body != expectedBody {
			t.Errorf("Expected truncated body, got %s", blob.Body)
		}
		if *blob.ContentLength != 50 {
			t.Errorf("Expected content length 50, got %d", *blob.ContentLength)
		}
	})
}

func TestZipArchive_ProcessFiles(t *testing.T) {
	t.Parallel()

	t.Run("empty archive", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		err := w.Close()
		if err != nil {
			t.Fatalf("Failed to close zip writer: %v", err)
		}

		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		var archive zip2jsons.ZipArchive = zip2jsons.ZipArchive{Reader: r}
		called := 0
		err = archive.ProcessFiles(func(file *zip.File) error {
			called++
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if called != 0 {
			t.Errorf("Handler called %d times, expected 0", called)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		f1, _ := w.Create("file1.txt")
		_, _ = f1.Write([]byte("content1"))
		f2, _ := w.Create("file2.json")
		_, _ = f2.Write([]byte("content2"))
		closeErr := w.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close zip writer: %v", closeErr)
		}
		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		var archive zip2jsons.ZipArchive = zip2jsons.ZipArchive{Reader: r}
		calledFiles := make(map[string]bool)
		err = archive.ProcessFiles(func(file *zip.File) error {
			calledFiles[file.Name] = true
			return nil
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if !calledFiles["file1.txt"] || !calledFiles["file2.json"] {
			t.Errorf("Expected both files to be processed, got %v", calledFiles)
		}
	})

	t.Run("handler returns error", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		f1, _ := w.Create("file1.txt")
		_, _ = f1.Write([]byte("content1"))
		f2, _ := w.Create("file2.json")
		_, _ = f2.Write([]byte("content2"))
		closeErr := w.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close zip writer: %v", closeErr)
		}

		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		var archive zip2jsons.ZipArchive = zip2jsons.ZipArchive{Reader: r}
		expectedErr := io.ErrClosedPipe // Any error
		err = archive.ProcessFiles(func(file *zip.File) error {
			if file.Name == "file2.json" {
				return expectedErr
			}
			return nil
		})
		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}
