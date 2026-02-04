gogogo: build run

build:
	@go build -tags debug -o bin/app ./app

run:
	@bin/app
