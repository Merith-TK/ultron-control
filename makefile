default:
	docker-compose up --remove-orphans
down:
	docker-compose down --remove-orphans

stop: down
restart: down default

pull:
	git pull