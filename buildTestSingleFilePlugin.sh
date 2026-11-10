#!/bin/sh
set -e

mkdir -p tmpPlugin
GOOS="wasip1" GOARCH="wasm" go build -C public/wasm/examples/uppercase -buildmode=c-shared -o ../../../../tmpPlugin/plugin.wasm  .
cp public/wasm/examples/uppercase/testdata/plugin.yaml ./tmpPlugin/plugin.yaml
tar cf uppercase.bpl -C tmpPlugin plugin.yaml plugin.wasm
rm -rf ./tmpPlugin

echo "Plugin Generated: ./uppercase.bpl"
