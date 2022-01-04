.PHONY: all
all:


.PHONY: build
build:
	@echo "===========> build binary"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./_output/wssvc ./cmd/apiserver.go

.PHONY: image
image:
	@$(MAKE) build
	@echo "===========> build image"
	@cd ./build/docker/ && ./build_image.sh

