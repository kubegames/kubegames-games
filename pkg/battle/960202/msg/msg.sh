#!/usr/bin/env bash
echo banker_niuniu
rm -f ./*.proto.go
protoc --go_out=plugins=grpc:. *.proto

#替换
grep -rl ',omitempty' ./*.pb.go | xargs sed -i "" "s/,omitempty//g"
