NAME = crawler

APP = cmd/app/main.go

BUILD = ./build


all: fmt clean build info

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD)

build:
	go build -o $(BUILD)/$(NAME) $(APP)

info:
	@echo
	@echo "$(BUILD)/$(NAME)"
	@echo "-markets=yandex, wb, ozon, avito (default "wb")"
	@echo "-query=Search query"
	@echo "-limit=		Top N results (default 10)"
	@echo "-page=		Page number (default 1)"
	@echo "-sort=		Sort by "popular" or "rating", etc (default "popular")"	
	@echo "-id=			Product ID (product info on the individual marketplace) (default "200135094")"
	@echo
	@echo "Example:"
	@echo "$(BUILD)/$(NAME) -markets="wb, ozon" -query="test" -limit=5"

.PHONY: all fmt clean build info