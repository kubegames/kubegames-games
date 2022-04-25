#!/usr/bin/env bash
echo 13水协议......
protoc --go_out=plugins=grpc:. *.proto

#替换
grep -rl ',omitempty' ./*.pb.go | xargs sed -i "" "s/,omitempty//g"

#rm -rf ~/Desktop/public_msg/13water/*.proto
#同时复制到给前段看的协议里面
#cp *.proto ~/Desktop/public_msg/13water/

#cd ~/Desktop/public_msg/;git add .;git commit -m 更新13水协议文档;git pull;git push
