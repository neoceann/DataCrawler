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
	@echo "-limit=		Top N results (default 5)"
	@echo "-sort=		Sort by rating, popular, priceAsc, priceDesc (default popular)"
	@echo
	@echo "Example:"
	@echo "$(BUILD)/$(NAME) -markets=wb, ozon -query=\""lenovo notepad\"" -limit=3"

.PHONY: all fmt clean build info