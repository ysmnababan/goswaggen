include ./Makefile.Common

.PHONY: test-all
test-all:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
