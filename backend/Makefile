APP_NAME = warehouse_management_system
MAIN_PATH = ./cmd
BUILD_DIR = ./bin
GOLANGCI_LINT = $(BUILD_DIR)/golangci-lint
GOLANGCI_LINT_VERSION = v1.62.2
GOTESTSUM = $(BUILD_DIR)/gotestsum

TEST_COMPOSE_FILE = docker-compose.test.yml
TEST_DB_DSN = "postgres://postgres:postgres@localhost:5431/test_db?sslmode=disable"

.PHONY: all
all: build

.PHONY: build
build: tidy
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

.PHONY: run
run:
	go run $(MAIN_PATH)

.PHONY: test
test: mock test-env-up .install-gotestsum
	TEST_DATABASE_URL=$(TEST_DB_DSN) $(GOTESTSUM) --format short-verbose -- ./... -count=1 || ($(MAKE) test-env-down && exit 1)
	@$(MAKE) test-env-down

.PHONY: test-env-up
test-env-up:
	docker compose -f $(TEST_COMPOSE_FILE) up -d

.PHONY: test-env-down
test-env-down:
	docker compose -f $(TEST_COMPOSE_FILE) down

.PHONY: mock
mock:
	go generate ./...

.PHONY: tidy
tidy: mock
	go fmt ./...
	go mod tidy

.PHONY: .install-linter
.install-linter:
	[ -f $(GOLANGCI_LINT) ] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BUILD_DIR) $(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: .install-linter
	$(GOLANGCI_LINT) run ./... --config=./.golangci.yml

.PHONY: .install-gotestsum
.install-gotestsum:
	[ -f $(GOTESTSUM) ] || GOBIN=$(abspath $(BUILD_DIR)) go install gotest.tools/gotestsum@latest
