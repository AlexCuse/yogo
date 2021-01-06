PHONY: test

GO=GO111MODULE=on go

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

test: test-scanner test-quoter

