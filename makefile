gogogo: build run

build:
	@go build -o bin/app app/*.go

run:
	@bin/app
