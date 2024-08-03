BIN="bin"
BINARY_NAME=flow

build:
	go build -o ./${BIN}/${BINARY_NAME}

run: build
	./${BIN}/${BINARY_NAME}

clean:
	go clean
	rm ./${BIN}/${BINARY_NAME}
