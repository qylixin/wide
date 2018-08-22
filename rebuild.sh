#!/bin/bash
set -x

rm wide.tar
docker build -t 88250/wide:latest .
docker save peersafe/wide:latest -o wide.tar
ssh peersafe@192.168.0.15 "cd ~/wide; rm wide.tar; docker-compose down"
scp wide.tar peersafe@192.168.0.15:~/wide
ssh peersafe@192.168.0.15 "cd ~/wide; docker load -i wide.tar; docker-compose up -d"
docker images | grep none| awk  '{ print $3 }'|xargs docker rmi
ssh peersafe@192.168.0.15 "docker images | grep none| awk  '{ print $3 }'|xargs docker rmi"
echo ok
