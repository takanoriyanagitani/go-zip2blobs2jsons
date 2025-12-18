#!/bin/sh

mkdir -p ./testdata.d/

printf helo | gzip --fast >./testdata.d/helo.txt.gz
printf wrld | gzip --fast >./testdata.d/wrld.txt.gz

find ./testdata.d -type f -name '*.gz' |
	zip \
		-@ \
		-T \
		-v \
		-o \
		./testdata.d/hw.zip

cat ./testdata.d/hw.zip |
	wazero \
		run \
		./zip2blobs2jsons.wasm \
		--zip-size-max 1048576 \
		--zip-name hw.zip \
		--item-size-max 131072 \
		--item-content-type text/plain \
		--item-content-encoding gzip |
	jq
