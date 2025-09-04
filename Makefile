.PHONY: build run docs clean

APP_NAME=dekamond-task
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) .

run:
	go run .

clean:
	rm -rf $(BUILD_DIR)

docs:
	swag init
