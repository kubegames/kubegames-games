#!/bin/bash
server_name="buyu-dev"
path="/home/buyu"
passwd="toor"
rm -rf  $server_name
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $server_name  main.go

/usr/bin/expect <<-EOF

set timeout 120  
spawn scp $server_name root@192.168.0.90:$path
expect {  
        "(yes/no)?" { send "yes\r" ; exp_continue }  
        "password:" { send "$passwd\r" }  
}  
expect "iBMC:/->"  


spawn scp -r config root@192.168.0.90:$path
expect {  
        "(yes/no)?" { send "yes\r" ; exp_continue }  
        "password:" { send "$passwd\r" }  
}  
expect "iBMC:/->"  
