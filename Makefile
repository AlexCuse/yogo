PHONY: test-go
test-go:
	cd go && make test fix build

PHONY: test-dashboard
test-dashboard:
	cd web/dashboard && npm install && npm run-script build

test: test-go test-dashboard

PHONY: docker-compose-build
docker-compose-build:
	docker-compose build

PHONY: docker-compose-run
docker-compose-run: docker-compose-build
	docker-compose up

