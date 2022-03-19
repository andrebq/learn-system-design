.PHONY: default test tidy watch dist serve-local install-statik generate

default: test

include Env.mk
-include EnvLocal.mk
include Docker.mk
include Stress.mk
include ControlPlane.mk
include Fleet.mk
include Support.mk

test:
	go test ./...

tidy:
	go mod tidy
	go fmt ./...

dist:
	go build -o ./dist/lsd ./cmd/lsd

serve-local: dist
	cp -Rv ./examples/* ./dist/examples/
	./dist/lsd serve \
		--bind 127.0.0.1:$(LOCAL_PORT) \
		--handler-file $(PWD)/examples/echo-handler/handler.lua \
		--public-endpoint http://127.0.0.1:$(LOCAL_PORT)


watch: generate
	modd -f modd.conf

generate:
	go generate ./...

install-statik:
	go install github.com/rakyll/statik@v0.1.7
