build:
	go build ./cmd/lunar

build-docker:
	docker build -t wendellsun/lunar .

update-docker:
	make build-docker
	docker push wendellsun/lunar
