build:
	go build .

test: build
	go test

cli:
	./chopsticks cli

build-and-cli: build
	$(MAKE) cli

serve:
	./chopsticks serve

build-and-serve: build
	$(MAKE) serve
