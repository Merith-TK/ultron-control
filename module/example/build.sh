#!/bin/bash
#	go build -buildmode=plugin -o workdir/modules/example.ult.so module/example/example.go
#	go build -buildmode=plugin -o workdir/modules/turtle.ult.so module/example/turtle.go

pushd module/example
for file in *.go
do 
    echo "Building $file"
    go build -buildmode=plugin -o ../../workdir/modules/$file.ult.so ./$file
done
popd