VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X github.com/paulrose/hatch/cmd.version=$(VERSION) \
	-X github.com/paulrose/hatch/cmd.commit=$(COMMIT) \
	-X github.com/paulrose/hatch/cmd.date=$(DATE)

.PHONY: build build-test run test clean app frontend icon

build:
	go build -ldflags '$(LDFLAGS)' -o hatch .

build-test:
	go build -ldflags '$(LDFLAGS)' -o testing/hatch .

run:
	go run -ldflags '$(LDFLAGS)' . $(ARGS)

test:
	go test ./...

frontend:
	cd frontend && npm install && npm run build

app: frontend
	MACOSX_DEPLOYMENT_TARGET=15.0 go build -ldflags '$(LDFLAGS)' -o hatch .

icon:
	rsvg-convert -f png -w 44 -h 44 split-solid-full.svg -o internal/tray/icon.png

clean:
	rm -f hatch
