.PHONY: help deps install-tools build build-linux run dev templ-generate templ-watch css-build css-watch deploy

BINARY=pilot
DEPLOY_HOST=claw
DEPLOY_PATH=~/workspaces/prastuvwxyz/pilot

help:
	@echo "Available commands:"
	@echo "  make deps            - Install Go dependencies"
	@echo "  make install-tools   - Verify dev tools"
	@echo "  make build           - Build binary"
	@echo "  make run             - Run binary"
	@echo "  make dev             - Hot reload (air + templ)"
	@echo "  make templ-generate  - Generate templ files"
	@echo "  make css-build       - Build TailwindCSS"
	@echo "  make css-watch       - Watch TailwindCSS"
	@echo "  make test            - Run tests"
	@echo "  make build-linux     - Cross-compile for Linux amd64"
	@echo "  make deploy          - Build linux binary + scp + restart on server"

deps:
	go mod download && go mod tidy

install-tools:
	@which templ       || (echo "❌ templ not found. Install: go install github.com/a-h/templ/cmd/templ@latest" && exit 1)
	@echo "✅ templ: $$(templ version)"
	@which air         || (echo "❌ air not found. Install: go install github.com/air-verse/air@latest" && exit 1)
	@echo "✅ air found"
	@which tailwindcss || (echo "❌ tailwindcss not found. Install: brew install tailwindcss" && exit 1)
	@echo "✅ tailwindcss found"

build: templ-generate css-build
	go build -o $(BINARY) ./cmd/web/

build-linux: templ-generate css-build
	GOOS=linux GOARCH=amd64 go build -o $(BINARY) ./cmd/web/

run: build
	./$(BINARY)

dev:
	air

templ-generate:
	templ generate

templ-watch:
	templ generate --watch

css-build:
	tailwindcss -i web/static/input.css -o web/static/output.css --minify

css-watch:
	tailwindcss -i web/static/input.css -o web/static/output.css --watch

test:
	go test ./... -v

deploy: build-linux
	scp $(BINARY) openclaw@$(DEPLOY_HOST):$(DEPLOY_PATH)/$(BINARY)
	scp web/static/output.css openclaw@$(DEPLOY_HOST):$(DEPLOY_PATH)/web/static/output.css
	ssh $(DEPLOY_HOST) "systemctl --user restart pilot"
	@echo "✅ deployed to $(DEPLOY_HOST)"
