APP_NAME=wwpm
BUILD_DIR=build

.PHONY: all clean build release

all: build

build:
	@echo "Building for linux/amd64 ..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "Building for linux/arm64 ..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) main.go

release: build
	tar czvf $(BUILD_DIR)/$(APP_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME)
	tar czvf $(BUILD_DIR)/$(APP_NAME)-linux-arm64.tar.gz -C $(BUILD_DIR) $(APP_NAME)

clean:
	rm -rf $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)