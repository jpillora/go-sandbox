#!/bin/sh

cd ../static/ &&
	echo "packing..." &&
	go-bindata -ignore="^src\/" -ignore="\/\."  -ignore="^\." -pkg=sandbox -o ../lib/files-prod.go -debug=false './...' &&
	echo "done"