GO15VENDOREXPERIMENT=1

NAME	 := exclamation
TARGET	 := $(NAME)
VERSION  := beta
DIST_DIRS := find * -type d -exec

SRCS	:= $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.version=$(VERSION)\" -extldflags \"-static\""

$(TARGET): $(SRCS)
	./makeTestConfig.sh
	./makeSSL.sh
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o $(NAME)

.PHONY: install wash clean cross-build upde dep dep-install dist

install:
	go install $(LDFLAGS)

wash:
	rm -f exclamation
	rm -rf ssl/
	rm -rf dist/*

clean: wash
	rm -rf vendor/*

cross-build: deps
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$$os-$$arch/$(NAME); \
		done; \
	done

upde:
	dep ensure -update

dep:
	dep ensure

dep-install:
	ifeq ($(shell command -v dep 2> /dev/null),)
		go get -u github.com/golang/dep/cmd/dep
	endif

dist:
	cd dist && \
		$(DIST_DIRS) cp ../LICENSE {} \; && \
		$(DIST_DIRS) cp ../README.md {} \; && \
		$(DIST_DIRS) tar -zcf $(NAME)-$(VERSION)-{}.tar.gz {} \; && \
		$(DIST_DIRS) zip -r $(NAME)-$(VERSION)-{}.zip {} \; && \
		cd ..
