#!/bin/bash

rm wide.tar.gz

mkdir wide

cp -r ../workspaces wide/
cp -r ../conf wide/
cp ../docker-compose.yaml wide/
cp ../README.md wide/
cp $GOPATH/src/github.com/peersafe/bcap/build/docker/build-api/LICENSE wide/

tar -zcvf wide.tar.gz wide

rm -rf wide