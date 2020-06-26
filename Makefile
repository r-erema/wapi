.PHONY: deploy build run

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

lint:
	golangci-lint run -v
