package zip2jsons

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"

	bj "github.com/takanoriyanagitani/go-blob2json"
)

// FileLike represents a file-like object with ReaderAt and Size.
type FileLike struct {
	io.ReaderAt

	Size int64
}

// ToZip converts a FileLike object into a ZipArchive.
func (l FileLike) ToZip() (ZipArchive, error) {
	rdr, e := zip.NewReader(l.ReaderAt, l.Size)
	if nil != e {
		return ZipArchive{}, fmt.Errorf("%w: %v", ErrNewReader, e)
	}
	return ZipArchive{Reader: rdr}, nil
}

// ByteReader wraps a bytes.Reader.
type ByteReader struct{ *bytes.Reader }

// AsFileLike converts a ByteReader to a FileLike object.
func (b ByteReader) AsFileLike() FileLike {
	return FileLike{
		ReaderAt: b.Reader,
		Size:     b.Size(),
	}
}

// Buffer wraps a bytes.Buffer.
type Buffer struct{ *bytes.Buffer }

// AsFileLike converts a Buffer to a FileLike object.
func (b Buffer) AsFileLike() FileLike {
	var dat []byte = b.Buffer.Bytes()
	var rdr *bytes.Reader = bytes.NewReader(dat)
	return ByteReader{Reader: rdr}.AsFileLike()
}

// ToFileLike converts a Buffer to a FileLike object by copying its content.
func (b Buffer) ToFileLike() FileLike {
	var dat []byte = b.Buffer.Bytes()
	var copied []byte = bytes.Clone(dat)
	var rdr *bytes.Reader = bytes.NewReader(copied)
	return ByteReader{Reader: rdr}.AsFileLike()
}

// Reader wraps an io.Reader.
type Reader struct{ io.Reader }

// ToLimited returns a new Reader that reads from r but stops after n bytes.
func (r Reader) ToLimited(limit int64) io.Reader {
	return io.LimitReader(r.Reader, limit)
}

// ToBuffer reads the content of the Reader up to a given limit into a Buffer.
func (r Reader) ToBuffer(limit int64) (Buffer, error) {
	var ltd io.Reader = r.ToLimited(limit)
	var buf bytes.Buffer
	_, e := io.Copy(&buf, ltd)
	if nil != e {
		return Buffer{}, fmt.Errorf("could not copy to buffer: %w", e)
	}
	return Buffer{Buffer: &buf}, nil
}

func (r Reader) toFileLike(limit int64) (FileLike, error) {
	buf, e := r.ToBuffer(limit)
	if nil != e {
		return FileLike{}, fmt.Errorf("could not convert to buffer: %w", e)
	}

	return buf.AsFileLike(), nil
}

func (r Reader) toZip(limit int64) (ZipArchive, error) {
	f, e := r.toFileLike(limit)
	if nil != e {
		return ZipArchive{}, fmt.Errorf("could not convert to file-like: %w", e)
	}

	arc, e := f.ToZip()
	if nil != e {
		return ZipArchive{}, fmt.Errorf("could not create zip archive: %w", e)
	}
	return arc, nil
}

// ToJsons reads a zip file from the Reader, processes its items into JSON blobs, and encodes them.
func (r Reader) ToJsons(limit int64, enc JsonEncoder, bldr bj.BlobBuilder) error {
	arc, e := r.toZip(limit)
	if nil != e {
		return fmt.Errorf("could not convert reader to zip archive: %w", e)
	}

	e = ProcessZipArchive(arc, enc, bldr)
	if nil != e {
		return fmt.Errorf("could not process zip archive: %w", e)
	}
	return nil
}
