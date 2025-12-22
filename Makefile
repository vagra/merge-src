BINARY_NAME=merge-src

all: build

build:
	go build -o $(BINARY_NAME) main.go

install:
	go install

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f *-analysis-*.txt

# 方便交叉编译 (例如在 Mac 上编译 Windows 版)
build-win:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) main.go