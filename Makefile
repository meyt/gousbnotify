.PHONY: run
run:
	go run .

.PHONY: build
build:
	go build -o dist/gousbnotify-linux-amd64

.PHONY: install
install:
	sudo apt-get install -y gcc libasound2-dev libudev-dev
	go mod tidy
