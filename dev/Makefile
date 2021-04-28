default: up

.PHONY: up
up:
	sh ./env.sh
	sh ./before-up.sh
	docker-compose up -d
	sh ./after-up.sh

.PHONY: down
down:
	sh ./env.sh
	docker-compose down

.PHONY: clean
clean:
	sh ./env.sh
	sh ./clean.sh
