PHONY: test

GO=GO111MODULE=on go

test-common:
	cd common && $(GO) test ./...
	cd common && $(GO) vet ./...
	gofmt -l common/
	[ "`gofmt -l common/`" = "" ]

test-scanner: 
	cd scanner  && $(GO) test ./...
	cd scanner && $(GO) vet ./...
	gofmt -l scanner/
	[ "`gofmt -l scanner/`" = "" ]

test-monitor:
	cd monitor && $(GO) test ./...
	cd monitor && $(GO) vet ./...
	gofmt -l monitor/
	[ "`gofmt -l monitor/`" = "" ]

test-quote-enricher:
	cd quote-enricher && $(GO) test ./...
	cd quote-enricher && $(GO) vet ./...
	gofmt -l quote-enricher/
	[ "`gofmt -l quote-enricher/`" = "" ]

test-history:
	cd history && $(GO) test ./...
	cd history && $(GO) vet ./...
	gofmt -l history/
	[ "`gofmt -l history/`" = "" ]

test-watch:
	cd watch && $(GO) test ./...
	cd watch && $(GO) vet ./...
	gofmt -l watch/
	[ "`gofmt -l watch/`" = "" ]

test-signals:
	cd signals && $(GO) test ./...
	cd signals && $(GO) vet ./...
	gofmt -l signals/
	[ "`gofmt -l signals/`" = "" ]

test-dashboard:
	cd dashboard && npm run-script build

test: test-common test-scanner test-monitor test-quote-enricher test-history test-watch test-signals test-dashboard

PHONY: docker-compose-build
docker-compose-build:
	docker-compose build --build-arg GITHUB_KEY

PHONY: docker-compose-run
docker-compose-run: docker-compose-build
	docker-compose up