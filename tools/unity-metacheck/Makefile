NAME    := unity-metacheck
SRCS    := $(shell find . -name "*.go" -type f)
LDFLAGS := -ldflags="-s -w -extldflags -static"

bin/$(NAME): $(SRCS)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/darwin-amd64/$(NAME) $(SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/linux-amd64/$(NAME) $(SRCS)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/windows-amd64/$(NAME).exe $(SRCS)

.PHONY: clean
clean:
	rm -rf bin/*
