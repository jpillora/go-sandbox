#!/bin/sh

cd ../ &&
	echo "packing..." &&
	go-bindata -ignore="\/src\/" -ignore="^src\/" -ignore="\/\."  -ignore="^\." -pkg=sandbox -o ./lib/files-prod.go -debug=false './static/...' &&
	echo "done"