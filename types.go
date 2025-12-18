package zip2jsons

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"time"

	bj "github.com/takanoriyanagitani/go-blob2json"
)

// ZipArchive wraps a zip.Reader for easier file processing.
type ZipArchive struct{ *zip.Reader }

// Files returns the list of files within the zip archive.
func (a ZipArchive) Files() []*zip.File { return a.Reader.File }

// ProcessFiles iterates through each file in the zip archive and applies the given handler.
func (a ZipArchive) ProcessFiles(
	handler func(*zip.File) error,
) error {
	var files []*zip.File = a.Files()
	for _, file := range files {
		e := handler(file)
		if nil != e {
			return fmt.Errorf("error processing file %s: %w", file.Name, e)
		}
	}
	return nil
}

// ZipItem represents a single file within a zip archive.
type ZipItem struct{ *zip.File }

// Header returns the FileHeader of the zip item.
func (i ZipItem) Header() zip.FileHeader { return i.File.FileHeader }

// Modified returns the modification time of the zip item.
func (i ZipItem) Modified() time.Time { return i.Header().Modified }

// Name returns the name of the zip item.
func (i ZipItem) Name() string { return i.Header().Name }

// ToBlob converts a ZipItem into a bj.Blob, applying content limits and base64 encoding.
func (i ZipItem) ToBlob(builder bj.BlobBuilder) (*bj.Blob, error) {
	bldr := builder
	var modified time.Time = i.Modified()
	bldr.LastModified = &modified

	rc, e := i.File.Open()
	if nil != e {
		return nil, fmt.Errorf("could not open zip file %s: %w", i.Name(), e)
	}
	defer rc.Close() //nolint:errcheck// the "file" is read only

	// if bldr.MaxBytes is unset(0; not initialized?), the blob will be empty
	return bldr.NewBlobFromReader(rc, i.Name())
}

// JsonEncoder wraps a json.Encoder for encoding blobs.
type JsonEncoder struct{ *json.Encoder }

// EncodeBlob encodes a bj.Blob to the underlying JSON encoder.
func (j JsonEncoder) EncodeBlob(b *bj.Blob) error {
	e := j.Encoder.Encode(b)
	if nil != e {
		return fmt.Errorf("could not encode blob: %w", e)
	}
	return nil
}
