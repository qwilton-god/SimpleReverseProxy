build:
	mkdir -p bin
	go build -o bin/proxy ./cmd/proxy
	go build -o bin/servers ./cmd/server

run-servers:
	go run ./cmd/server

run-proxy:
	go run ./cmd/proxy

clean:
	rm -rf bin/
