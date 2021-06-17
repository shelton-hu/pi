default: up

.PHONY: up
up: check
	sh ./env.sh
	sh ./before-start.sh
	docker-compose up -d
	sh ./after-start.sh

.PHONY: down
down: check
	docker-compose down

.PHONY: restart
restart: check
	make down
	make up

.PHONY: put-config
put-config: check
	sh ./after-start.sh

.PHONY: check
check:
	sh ./check.sh
