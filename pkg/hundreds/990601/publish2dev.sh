GOOS=linux GOARCH=amd64 go build -o birdanimal

## config
# scp ./config/birdanimal.json root@192.168.0.46:/home/servers/birdanimal/config/
# scp ./config/game_frame.json root@192.168.0.46:/home/servers/birdanimal/config/
# scp ./config/robot.json root@192.168.0.46:/home/servers/birdanimal/config/

## project
scp ./birdanimal root@192.168.0.46:/home/servers/birdanimal
