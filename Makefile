.PHONY: deploy build run lint

deploy: build
	docker push erema/wapi:0.1

build:
	docker build . -t erema/wapi:0.1

run:
	docker run --env-file .env erema/wapi:0.1

gen-mock:
	mockgen  -source=internal/service/auth/auth.go -destination=internal/testutil/mock/auth/auth.go
	mockgen  -source=internal/service/listener/listener.go -destination=internal/testutil/mock/listener/listener.go
	mockgen  -source=internal/repository/session/repository.go -destination=internal/testutil/mock/session/repository.go

GOLANGCI_LINT_VERSION := v1.27.0
GOLANGCI_LINT_PATH := $(shell go env GOPATH)/bin/golangci-lint-versions/$(GOLANGCI_LINT_VERSION)
lint:
ifeq ("$(wildcard $(GOLANGCI_LINT_PATH))","")
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
	| sh -s -- -b $(GOLANGCI_LINT_PATH) $(GOLANGCI_LINT_VERSION)
endif
	$(GOLANGCI_LINT_PATH)/golangci-lint run -v
