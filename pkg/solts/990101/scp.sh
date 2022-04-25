#!/bin/bash
server_name="sgxml-dev"
path="/home/sgxml-server"
passwd="toor"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $server_name  sgxml.go

/usr/bin/expect <<-EOF

set timeout 120  
spawn scp $server_name root@192.168.0.90:$path
expect {  
        "(yes/no)?" { send "yes\r" ; exp_continue }  
        "password:" { send "$passwd\r" }  
}  
expect "iBMC:/->"  


spawn scp -r conf root@192.168.0.90:$path
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