ENTRYPOINT=./cmd/klevliga
SERVICE_NAME=klevliga
GO_VERSION=1.26.0
COMPOSE=docker compose -f ../../infra/docker-compose.yml
WEB_DIR=../../apps/web

.PHONY: lint
lint:
	golangci-lint run

.PHONY: build
build: tidy
	CGO_ENABLED=0 go build -o ./bin/${SERVICE_NAME} ${ENTRYPOINT}

.PHONY: tidy
tidy:
	@go mod tidy -go ${GO_VERSION}

.PHONY: gen
gen:
	go install tool
	go generate ./...
	go tool sqlc generate

.PHONY: test
test: tidy test-unit

.PHONY: test-unit
test-unit:
	go test -race ./...

.PHONY: test-integration
test-integration:
	go test -race -tags=integration ./...

.PHONY: run
run:
	go run ${ENTRYPOINT} -c ./config.yaml

.PHONY: mail-check
mail-check:
	go run ./cmd/mailcheck -c ./config.yaml

.PHONY: ci-build
ci-build:
	CGO_ENABLED=0 go build \
	-o ./bin/${SERVICE_NAME} \
	-ldflags "-X 'main.version=${VERSION}' -X 'main.buildTime=${BUILD_TIME}'" \
	${ENTRYPOINT}

.PHONY: db-up
db-up:
	$(COMPOSE) up -d

.PHONY: db-down
db-down:
	$(COMPOSE) down

.PHONY: web-install
web-install:
	cd $(WEB_DIR) && npm install

.PHONY: web
web:
	cd $(WEB_DIR) && npm run dev
