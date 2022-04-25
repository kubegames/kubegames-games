#！/bin/sh
#-----------------------------------
#-----------------------------------


# go root的路径
_GoLangRootDir="/home/day/MyWork/Paths/Go/Root/go_1_12_13_x64"
# go path路径
_GoPath="/home/day/MyWork/Paths/Go/Path"
# 编译平台
_GoPlatform="amd64"
# 编译系统版本
_GoOsVer="linux"



echo "================== 开始设置go环境 =================="
export GOROOT=$_GoLangRootDir
export GOPATH=$_GoPath
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
export GO111MODULE=$_GoModule
export GOHOSTARCH=$_GoPlatform
export GOHOSTOS=$_GoOsVer
export GONOPROXY=$_GoProxy
export CGO_ENABLED=$_GoCGOEnabled
echo "================== 设置go环境完成 =================="
echo "================== go的版本是: =================="
go version


echo "==================  =================="
protoc --go_out=plugins=grpc:. erbagang.proto
