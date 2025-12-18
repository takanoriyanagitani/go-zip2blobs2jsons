package zip2jsons_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	bj "github.com/takanoriyanagitani/go-blob2json"
	"github.com/takanoriyanagitani/go-zip2blobs2jsons"
)

func TestProcessZipArchive(t *testing.T) {
	t.Parallel()

	t.Run("empty zip", func(t *testing.T) {
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
		var enc zip2jsons.JsonEncoder = zip2jsons.JsonEncoder{Encoder: json.NewEncoder(new(bytes.Buffer))}
		var bldr bj.BlobBuilder = bj.BlobBuilder{MaxBytes: 1024}

		err = zip2jsons.ProcessZipArchive(archive, enc, bldr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("single file", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		f, err := w.Create("test.txt")
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write([]byte("hello world"))
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

		var archive zip2jsons.ZipArchive = zip2jsons.ZipArchive{Reader: r}
		outBuf := new(bytes.Buffer)
		var enc zip2jsons.JsonEncoder = zip2jsons.JsonEncoder{Encoder: json.NewEncoder(outBuf)}
		var bldr bj.BlobBuilder = bj.BlobBuilder{
			ContentType: "text/plain",
			MaxBytes:    1024,
		}

		err = zip2jsons.ProcessZipArchive(archive, enc, bldr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		var blob bj.Blob
		err = json.NewDecoder(outBuf).Decode(&blob)
		if err != nil {
			t.Fatalf("Failed to decode JSON output: %v", err)
		}

		if blob.Name != "test.txt" {
			t.Errorf("Expected name 'test.txt', got '%s'", blob.Name)
		}
		if blob.Body != "aGVsbG8gd29ybGQ=" { // base64 of "hello world"
			t.Errorf("Expected body 'aGVsbG8gd29ybGQ=', got '%s'", blob.Body)
		}
		if blob.ContentType != "text/plain" {
			t.Errorf("Expected content type 'text/plain', got '%s'", blob.ContentType)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		f1, err := w.Create("file1.txt")
		if err != nil {
			t.Fatalf("Failed to create file1: %v", err)
		}
		_, err = f1.Write([]byte("content1"))
		if err != nil {
			t.Fatalf("Failed to write to file1: %v", err)
		}

		f2, err := w.Create("file2.json")
		if err != nil {
			t.Fatalf("Failed to create file2: %v", err)
		}
		_, err = f2.Write([]byte(`{"key": "value"}`))
		if err != nil {
			t.Fatalf("Failed to write to file2: %v", err)
		}
		closeErr := w.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close zip writer: %v", closeErr)
		}
		r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			t.Fatalf("Failed to create zip reader: %v", err)
		}

		var archive zip2jsons.ZipArchive = zip2jsons.ZipArchive{Reader: r}
		outBuf := new(bytes.Buffer)
		var enc zip2jsons.JsonEncoder = zip2jsons.JsonEncoder{Encoder: json.NewEncoder(outBuf)}
		var bldr bj.BlobBuilder = bj.BlobBuilder{
			ContentType: "application/octet-stream", // Default
			MaxBytes:    1024,
		}

		err = zip2jsons.ProcessZipArchive(archive, enc, bldr)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode multiple JSON objects from the output
		decoder := json.NewDecoder(outBuf)
		var blobs []bj.Blob
		for {
			var blob bj.Blob
			err := decoder.Decode(&blob)
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("Failed to decode JSON output: %v", err)
			}
			blobs = append(blobs, blob)
		}

		if len(blobs) != 2 {
			t.Fatalf("Expected 2 blobs, got %d", len(blobs))
		}

		// Check blob 1
		if blobs[0].Name != "file1.txt" {
			t.Errorf("Expected name 'file1.txt', got '%s'", blobs[0].Name)
		}
		if blobs[0].Body != "Y29udGVudDE=" { // base64 of "content1"
			t.Errorf("Expected body 'Y29udGVudDE=', got '%s'", blobs[0].Body)
		}

		// Check blob 2
		if blobs[1].Name != "file2.json" {
			t.Errorf("Expected name 'file2.json', got '%s'", blobs[1].Name)
		}
		if blobs[1].Body != "eyJrZXkiOiAidmFsdWUifQ==" { // base64 of `{"key": "value"}`
			t.Errorf("Expected body 'eyJrZXkiOiAidmFsdWUifQ==', got '%s'", blobs[1].Body)
		}
	})
}
