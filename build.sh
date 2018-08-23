#!/bin/bash
set -x

rm wide.tar wide.tar.gz
docker rmi -f peersafe/wide:latest
GOOS=linux GOARCH=amd64 go build  -tags "nopkcs11"
docker build --no-cache -t peersafe/wide:latest .
#docker save peersafe/wide:latest -o wide.tar
#tar -zcvf wide.tar.gz wide.tar
#ssh peersafe@192.168.0.15 "cd ~/wide; rm wide.tar; docker-compose down"
#scp wide.tar peersafe@192.168.0.15:~/wide
#ssh peersafe@192.168.0.15 "cd ~/wide; docker load -i wide.tar; docker-compose up -d"
#docker images | grep none| awk  '{ print $3 }'|xargs docker rmi
#ssh peersafe@192.168.0.15 "docker images | grep none| awk  '{ print $3 }'|xargs docker rmi"
echo ok
