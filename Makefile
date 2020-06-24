.PHONY: deploy build run

deploy: build
	docker push erema/wapi:0.1

build:
	docker build . -t erema/wapi:0.1

run:
	docker run --env-file .env erema/wapi:0.1

gen-mock:
	mockgen  -source=src/service/auth/auth.go -destination=testutil/mock/auth/auth.go
	mockgen  -source=src/service/listener/listener.go -destination=testutil/mock/listener/listener.go
	mockgen  -source=src/repository/session/session.go -destination=testutil/mock/session/session.go

lint:
	golangci-lint run
