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

test-quoter:
	cd quoter && $(GO) test ./...
	cd quoter && $(GO) vet ./...
	gofmt -l quoter/
	[ "`gofmt -l quoter/`" = "" ]

test-quote-enricher:
	cd quote-enricher && $(GO) test ./...
	cd quote-enricher && $(GO) vet ./...
	gofmt -l quote-enricher/
	[ "`gofmt -l quote-enricher/`" = "" ]

test: test-common test-scanner test-quoter test-quote-enricher

