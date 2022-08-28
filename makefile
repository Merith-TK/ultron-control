default:
	mkdir -p workdir/modules
	docker-compose up --remove-orphans
down:
	docker-compose down --remove-orphans

stop: down
restart: down default

air-setup:
	bash module/example/build.sh
	go build -o tmp/main.exe

pull:
	git pull

example-plugin:
	@echo "Building Example Plugin to workdir/modules"
	go build -buildmode=plugin -o workdir/modules/hello.ult.so module/example/hello.go