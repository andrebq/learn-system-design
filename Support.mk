.PHONY: jaeger-up jaeger-down


jaeger-up: dist
	./dist/lsd support jaeger docker-compose | tee ./localfiles/jaeger-compose.yaml
	docker-compose -f ./localfiles/jaeger-compose.yaml up

jaeger-down:
	docker-compose -f ./localfiles/jaeger-compose.yaml down
