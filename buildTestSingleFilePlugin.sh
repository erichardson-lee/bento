#!/bin/sh
set -e

# DEVELOPMENT FILE DO NOT MERGE TO MAIN

# 1. Make Temporary Plugin Folder
mkdir -p ./tmp/plugin ./target/plugins

# 2. Compile the uppercase WASM example into that folder
GOOS="wasip1" GOARCH="wasm" go build -C public/wasm/examples/uppercase -buildmode=c-shared -o ../../../../tmp/plugin/plugin.wasm  .

# 3. Copy the example plugin.yaml into that folder
cp public/wasm/examples/uppercase/testdata/plugin.yaml ./tmp/plugin/plugin.yaml

# 4. Make a tar archive containing both files
tar cf ./target/plugins/uppercase.bpl -C ./tmp/plugin plugin.yaml plugin.wasm

# 5. Remove the temporary folder
rm -rf ./tmp/plugin

echo "Plugin Generated: ./target/plugins/uppercase.bpl"
