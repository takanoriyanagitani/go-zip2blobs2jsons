package zip2jsons

import (
	"archive/zip"
	"fmt"

	bj "github.com/takanoriyanagitani/go-blob2json"
)

// BlobBuilder is a type alias for bj.BlobBuilder.
type BlobBuilder = bj.BlobBuilder

// ZipBlobsBuilder extends bj.BlobBuilder with a ZipName field.
type ZipBlobsBuilder struct {
	bj.BlobBuilder

	ZipName string
}

// ToBuilder converts a ZipBlobsBuilder to a bj.BlobBuilder.
func (z ZipBlobsBuilder) ToBuilder() bj.BlobBuilder {
	return bj.BlobBuilder{
		ContentType:     z.BlobBuilder.ContentType,
		ContentEncoding: z.BlobBuilder.ContentEncoding,
		MaxBytes:        z.BlobBuilder.MaxBytes,
		Metadata:        map[string]string{"ZipName": z.ZipName},
		LastModified:    z.BlobBuilder.LastModified,
	}
}

// ProcessZipArchive processes the files within a ZipArchive, converts each to a Blob, and encodes them to JSON.
func ProcessZipArchive(arc ZipArchive, enc JsonEncoder, bldr bj.BlobBuilder) error {
	e := arc.ProcessFiles(
		func(zfile *zip.File) error {
			zitem := ZipItem{File: zfile}
			blb, e := zitem.ToBlob(bldr)
			if nil != e {
				return fmt.Errorf("could not convert zip item to blob: %w", e)
			}
			e = enc.EncodeBlob(blb)
			if nil != e {
				return fmt.Errorf("could not encode blob: %w", e)
			}
			return nil
		},
	)
	if nil != e {
		return fmt.Errorf("could not process zip files: %w", e)
	}
	return nil
}
