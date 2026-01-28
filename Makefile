build:
	mkdir -p bin
	cd proxy && go build -o ../bin/proxy .
	cd server && go build -o ../bin/servers .

run-servers:
	cd server && go run .

run-proxy:
	cd proxy && go run .
