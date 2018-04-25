.PHONY: install all

all: install

install:
	go install github.com/cosmos/academy/go-examples/counter/.

build:
	go build -o build/counter github.com/cosmos/academy/go-examples/counter/.

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install

clean:
	rm -rf vendor/
	rm -rf build/
