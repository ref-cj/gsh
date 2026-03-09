gogogo: build run

build:
	@go build -o bin/release/app ./app # do a release build so see if it builds
	@go build -tags debug -o bin/debug/app ./app

run:
	@bin/debug/app
