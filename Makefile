VERSION := 0.1.1
GLIDE_COMMIT := 91d42a717b7202c55568a7da05be915488253b8d
LINTER_COMMIT := 052c5941f855d3ffc9e8e8c446e0c0a8f0445410

build:
	go build -ldflags="-X main.version=$(VERSION)" -o collectd-docker-plugin main.go

deps:
	@if [ "$$(which glide)" = "" ]; then \
		go get -v github.com/Masterminds/glide; \
		cd $$GOPATH/src/github.com/Masterminds/glide;\
		git checkout $(GLIDE_COMMIT);\
		go install;\
	fi
	glide install
	go install
	glide install

dist:
	which gox && echo "" || go get github.com/mitchellh/gox
	rm -rf tmp dist
	gox -os="linux windows freebsd openbsd" -output='tmp/{{.OS}}-{{.Arch}}-$(VERSION)/{{.Dir}}' -ldflags="-X main.version=$(VERSION)"
	mkdir dist

	# Create archives for Windows
	@for i in $$(find ./tmp -type f -name "collectd-docker-plugin.exe" | awk -F'/' '{print $$3}'); \
	do \
	  zip -j "dist/collectd-docker-plugin-$$i.zip" "./tmp/$$i/collectd-docker-plugin.exe"; \
	done

	# Create achrives for everything else
	@for i in $$(find ./tmp -type f -not -name "collectd-docker-plugin.exe" | awk -F'/' '{print $$3}'); \
	do \
	  chmod +x "./tmp/$$i/collectd-docker-plugin"; \
	  tar -zcvf "dist/collectd-docker-plugin-$$i.tar.gz" --directory="./tmp/$$i" "./collectd-docker-plugin"; \
	done

	rm -rf tmp

install: deps test
	go install -ldflags="-X main.version=$(VERSION)" main.go

setup-linter:
	@if [ "$$(which gometalinter)" = "" ]; then \
		go get -v github.com/alecthomas/gometalinter; \
		cd $$GOPATH/src/github.com/alecthomas/gometalinter;\
		git checkout $(LINTER_COMMIT);\
		go install;\
		gometalinter --install;\
	fi

test:
	make setup-linter
	gometalinter --vendor --fast --dupl-threshold=100 --cyclo-over=25 --min-occurrences=5 --disable=gas ./...
