build:
	cd src && go build .

test: build
	cd src && go test

cli:
	./src/chopsticks cli

build-and-cli: build
	$(MAKE) cli

serve:
	./src/chopsticks serve

build-and-serve: build
	$(MAKE) serve
