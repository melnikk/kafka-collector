default: run

prepare:
	go get github.com/cyberdelia/go-metrics-graphite

run: prepare
	@go run main.go
