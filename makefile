.PHONY: gogogo clean build rundev  rundevlog runrel
gogogo: build rundev

clean: 
	@rm -r bin/

build:
	@go build -o bin/release/app ./app & # do a release build to see if it builds (sometimes broken builds are hard to detect because of build tags)
	@go build -tags debug -o bin/debug/app ./app

rundev:
	@bin/debug/app
rundevlog:
	@bin/debug/app 2>log
runrel:
	@bin/release/app
