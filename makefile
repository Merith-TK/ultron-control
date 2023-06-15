default:
	mkdir -p workdir/modules
	docker-compose up --remove-orphans
down:
	docker-compose down --remove-orphans

stop: down
restart: down default

air:
	go build -buildvcs=false -o tmp/main.exe

air-setup:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

run: 
	go build -o ult.exe .
	./ult.exe

pull:
	git stash
	git pull
	git stash pop
