.PHONY: gogogo clean build run runrel
gogogo: build run

clean: 
	@rm -r bin/

build:
	@go build -o bin/release/app ./app & # do a release build to see if it builds (sometimes broken builds are hard to detect because of build tags)
	@go build -tags debug -o bin/debug/app ./app

run:
	@bin/debug/app
runrel:
	@bin/release/app
