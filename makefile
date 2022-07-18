default:
	mkdir -p workdir
	UID=${UID} GID=${GID} docker-compose up --remove-orphans
down:
	docker-compose down --remove-orphans

stop: down
restart: down default

pull:
	git pull

example-plugin:
	@echo "Building Example Plugin to workdir/modules"
	go build -buildmode=plugin -o workdir/modules/hello.ult.so module/example/hello.go