#!/bin/bash
server_name="birdAnimal-dev"
path="/home/devcd/deploygames/birdAnimal"
passwd="en6xXFQrLz"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags shanghai -o $server_name  main.go 

/usr/bin/expect <<-EOF

set timeout 120  
spawn scp $server_name devcd@192.168.11.62:$path
expect {  
        "(yes/no)?" { send "yes\r" ; exp_continue }  
        "password:" { send "$passwd\r" }  
}  
expect "iBMC:/->"  