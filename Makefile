GOBIN = $(shell which go)

TARGET := wordle

all: fmt vet $(TARGET)

fmt:
	$(GOBIN) fmt ./...

vet:
	$(GOBIN) vet ./...

$(TARGET): fmt vet
	$(GOBIN) build -o bin/$(TARGET) ./cmd/wordle
