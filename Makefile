export APP_NAME=app
export BIN_DIR=bin

.PHONY: build
build: clean
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -ldflags '-s -w' -o ${BIN_DIR}/${APP_NAME} main.go

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: deploy
deploy: build
	sls deploy -f app