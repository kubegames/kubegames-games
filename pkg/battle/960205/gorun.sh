#！/bin/sh
#-----------------------------------
# go 编译运行脚本
#-----------------------------------

# go root的路径
_GoLangRootDir="/home/day/MyWork/Paths/Go/root/go_1_12_13_x64"
# go path路径
_GoPath="/home/day/MyWork/Paths/Go/path"
# 编译平台
_GoPlatform="amd64"
# 编译系统版本
_GoOsVer="linux"
# go module代理
_GoProxy="https://athens.azurefd.net/"
# module 模式
_GoModule=auto
# 一些cgo编译的配置
_GoCGOEnabled="1"


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


echo "================== 开始编译erbagang =================="
go build -o erbagang main.go
./erbagang
