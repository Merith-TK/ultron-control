default:
	mkdir -p workdir/modules
	docker-compose up --remove-orphans
down:
	docker-compose down --remove-orphans

stop: down
restart: down default

air:
	bash module/example/build.sh
	go build -o tmp/main.exe

air-setup:
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

pull:
	git stash
	git pull
	git stash pop

example-plugin:
	@echo "Building Example Plugin to workdir/modules"
	go build -buildmode=plugin -o workdir/modules/hello.ult.so module/example/hello.go