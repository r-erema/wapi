.PHONY: deploy build run lint

GOLANGCI_LINT_VERSION := v1.27.0
GOLANGCI_LINT_PATH := $(shell go env GOPATH)/bin/golangci-lint-versions/$(GOLANGCI_LINT_VERSION)

deploy: build
	docker push erema/wapi:0.1

build:
	docker build . -t erema/wapi:0.1

run:
	docker run --env-file .env erema/wapi:0.1

# https://github.com/golang/mock
gen-mock:
	mockgen -package="mock" -source=internal/service/auth.go -destination=internal/testutil/mock/auth.go
	mockgen -package="mock" -source=internal/service/listener.go -destination=internal/testutil/mock/listener.go
	mockgen -package="mock" -source=internal/repository/repository.go -destination=internal/testutil/mock/repository.go
	mockgen -package="mock" -source=internal/service/supervisor.go -destination=internal/testutil/mock/connection.go
	mockgen -package="mock" -source=internal/infrastructure/http/client.go -destination=internal/testutil/mock/client.go
	mockgen -package="mock" -source=internal/infrastructure/whatsapp/conn.go -destination=internal/testutil/mock/conn.go

lint:
ifeq ("$(wildcard $(GOLANGCI_LINT_PATH))","")
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
	| sh -s -- -b $(GOLANGCI_LINT_PATH) $(GOLANGCI_LINT_VERSION)
endif
	$(GOLANGCI_LINT_PATH)/golangci-lint run -v

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
