NAME=ALL
PUSH=false
DEPLOY=true

all: hall web slots fishing dockerclean

slots:
	cd ./cmd/slots && ./build.sh $(NAME) $(PUSH)

fishing:
	cd ./cmd/fishing && ./build.sh $(NAME) $(PUSH)

hundreds:
	cd ./cmd/hundreds && ./build.sh $(NAME) $(PUSH)

battle:
	cd ./cmd/battle && ./build.sh $(NAME) $(PUSH)

web:
	cd ./client/web && ./build.sh $(PUSH) $(DEPLOY) 

dockerclean:
	docker images|grep none|awk '{print $3 }'|xargs docker rmi
