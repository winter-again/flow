bin_path = ./bin
bin_name = flow

.PHONY: try
try:
	go test -test.run=TestMain -v ./...

.PHONY: build
build:
	go build -o=${bin_path}/${bin_name} ${main_pkg_path}

.PHONY: run
run: build
	${bin_path}/${bin_name}

.PHONY: clean
clean:
	go clean
	rm ${bin_path}/${bin_name}

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: vet
vet:
	go vet

.PHONY: test
test:
	# go test -v ./...
	go test -skip TestMain -v ./...
