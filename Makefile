build:
	go build .

test: build
	go test

run:
	./chopsticks

build-and-run: build
	$(MAKE) run
