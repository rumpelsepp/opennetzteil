GO ?= go

all: netzteil netzteild

netzteil:
	$(GO) build $(GOFLAGS) -o $@ ./bin/$@/...

netzteild:
	$(GO) build $(GOFLAGS) -o $@ ./bin/$@/...

update:
	$(GO) get -u ./bin/...
	$(GO) mod tidy

clean:
	$(RM) netzteil
	$(RM) netzteild

.PHONY: all netzteil netzteild update clean
