deploy: build
	docker push erema/wapi:0.1

build:
	docker build . -t erema/wapi:0.1

run:
	docker run erema/wapi:0.1
