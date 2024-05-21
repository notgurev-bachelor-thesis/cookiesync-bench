.PHONY: remote deploy build

remote:
	ssh root@spb-w3-stathandler.moevideo.net

deploy: build
	scp -r * root@spb-w3-stathandler.moevideo.net:~/bench

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bench
