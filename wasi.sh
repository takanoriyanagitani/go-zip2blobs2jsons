#!/bin/sh

tinygo \
	build \
	-o ./zip2blobs2jsons.wasm \
	-target=wasip1 \
	-opt=z \
	-no-debug \
	./cmd/zip2blobs2jsons/main.go
