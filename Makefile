BINARY := battui
GOFLAGS := -ldflags="-s -w"

.PHONY: all build clean install uninstall run test vet lint

all: build

build:
	go build $(GOFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)

install: build
	install -d $(DESTDIR)/usr/local/bin
	install -m 755 $(BINARY) $(DESTDIR)/usr/local/bin/$(BINARY)

uninstall:
	rm -f $(DESTDIR)/usr/local/bin/$(BINARY)

run: build
	./$(BINARY)

test:
	go test -v -count=1 ./...

vet:
	go vet ./...

lint:
	@which staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"
