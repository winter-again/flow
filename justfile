bin_path := "./bin"
bin_name := "flow"

build:
    go build -o={{ bin_path }}/{{ bin_name }}

run: build
    {{ bin_path }}/{{ bin_name }}

clean:
    go clean
    rm {{ bin_path }}/{{ bin_name }}

test:
    go test -skip TestMain -v ./...
    # go test -v ./...

try:
    go test -test.run=TestMain -v ./...

tidy:
    go mod tidy -v

vet:
    go vet

