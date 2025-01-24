include .env

build:
	docker compose build

test:
	go test ./... -v -cover

run:
	docker compose up -d
