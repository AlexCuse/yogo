PHONY: test

GO=GO111MODULE=on go

test-scanner: 
	cd scanner  && $(GO) test ./...
	cd scanner && $(GO) vet ./...
	gofmt -l scanner/
	[ "`gofmt -l scanner/`" = "" ]

test: test-scanner

