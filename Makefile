build:
	go build .

test: build
	go test

run: test
	./chopsticks

build-and-run: test
	$(MAKE) run
