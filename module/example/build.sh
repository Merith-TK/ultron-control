#!/bin/bash

# This script is used for mass building modules in place if needbe

pushd module/example
for folder in */
do 
    echo "Building $folder"
	# remove trailing / from folder name
	folder2=${folder%/}
    go build -buildmode=plugin -o ../../workdir/modules/$folder2.ult.so ./$folder
done
popd