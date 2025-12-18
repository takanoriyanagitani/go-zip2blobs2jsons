package zip2jsons_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	bj "github.com/takanoriyanagitani/go-blob2json"
	"github.com/takanoriyanagitani/go-zip2blobs2jsons"
)

func TestReader_ToJsons_ZipSizeLimit(t *testing.T) {
	t.Parallel()

	t.Run("zip file exceeds limit", func(t *testing.T) {
		t.Parallel()

		// Create a zip file larger than the limit
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		f, err := w.Create("large.txt")
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write([]byte(strings.Repeat("a", 2000))) // 2KB content
		if err != nil {
			t.Fatalf("Failed to write to file in zip: %v", err)
		}
		closeErr := w.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close zip writer: %v", closeErr)
		}

		// Input reader for the ToJsons function
		var rdr zip2jsons.Reader = zip2jsons.Reader{Reader: bytes.NewReader(buf.Bytes())}
		var enc zip2jsons.JsonEncoder = zip2jsons.JsonEncoder{Encoder: json.NewEncoder(new(bytes.Buffer))}
		var bldr bj.BlobBuilder = bj.BlobBuilder{}

		// Set a limit smaller than the zip file
		zipSizeLimit := int64(100) // 100 bytes limit

		err = rdr.ToJsons(zipSizeLimit, enc, bldr)
		if err == nil {
			t.Errorf("Expected an error for zip file exceeding limit, got nil")
		}
		// Only checking that an error is returned is sufficient here.
	})

	t.Run("zip file within limit", func(t *testing.T) {
		t.Parallel()

		// Create a zip file smaller than the limit
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		f, err := w.Create("small.txt")
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write([]byte("short content")) // ~13 bytes content
		if err != nil {
			t.Fatalf("Failed to write to file in zip: %v", err)
		}
		closeErr := w.Close()
		if closeErr != nil {
			t.Fatalf("Failed to close zip writer: %v", closeErr)
		}

		// Input reader for the ToJsons function
		var rdr zip2jsons.Reader = zip2jsons.Reader{Reader: bytes.NewReader(buf.Bytes())}
		var enc zip2jsons.JsonEncoder = zip2jsons.JsonEncoder{Encoder: json.NewEncoder(new(bytes.Buffer))}
		var bldr bj.BlobBuilder = bj.BlobBuilder{}

		// Set a limit larger than the zip file
		zipSizeLimit := int64(1000) // 1KB limit

		err = rdr.ToJsons(zipSizeLimit, enc, bldr)
		if err != nil {
			t.Errorf("Unexpected error for zip file within limit: %v", err)
		}
		// Further checks could be added here to ensure correct output
	})
}
