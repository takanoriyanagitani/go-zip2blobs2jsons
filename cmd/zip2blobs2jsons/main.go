package main

import (
	"encoding/json"
	"flag"
	"os"

	zj "github.com/takanoriyanagitani/go-zip2blobs2jsons"
)

func main() {
	var zipSizeMax int64
	var zipName string
	var itemSizeMax int64
	var itemContentType string
	var itemContentEncoding string

	flag.Int64Var(&zipSizeMax, "zip-size-max", 10485760, "zip file size limit")
	flag.StringVar(&zipName, "zip-name", "unknown.zip", "zip file name")
	flag.Int64Var(&itemSizeMax, "item-size-max", 1048576, "zip item size limit")
	flag.StringVar(&itemContentType, "item-content-type", "application/octet-stream", "item content type")
	flag.StringVar(&itemContentEncoding, "item-content-encoding", "identity", "item content encoding")
	flag.Parse()

	var reader zj.Reader = zj.Reader{
		Reader: os.Stdin,
	}

	var encoder zj.JsonEncoder = zj.JsonEncoder{
		Encoder: json.NewEncoder(os.Stdout),
	}

	var builder zj.ZipBlobsBuilder = zj.ZipBlobsBuilder{
		ZipName: zipName,
	}
	builder.MaxBytes = itemSizeMax
	builder.ContentType = itemContentType
	builder.ContentEncoding = itemContentEncoding

	var bldr zj.BlobBuilder = builder.ToBuilder()

	e := reader.ToJsons(zipSizeMax, encoder, bldr)
	if nil != e {
		panic(e)
	}
}
