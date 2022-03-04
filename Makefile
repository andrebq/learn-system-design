.PHONY: default test tidy watch dist serve-local

default: test

include Env.mk
-include EnvLocal.mk
include Docker.mk

test:
	go test ./...

tidy:
	go mod tidy
	go fmt ./...

dist:
	go build -o ./dist/learn-system-design ./cmd/lsd

serve-local: dist
	./dist/learn-system-design serve --bind 127.0.0.1:$(LOCAL_PORT)


watch:
	modd -f modd.conf
