.PHONY: jaeger-up jaeger-down


jaeger-up: dist
	cd localfiles && ../dist/lsd support jaeger docker-compose
	docker-compose -f ./localfiles/jaeger.compose.yaml up

jaeger-down:
	docker-compose -f ./localfiles/jaeger.compose.yaml down


uptrace-up:
	cd localfiles && ../dist/lsd support uptrace docker-compose
	docker-compose -f ./localfiles/uptrace.compose.yaml up
uptrace-down:
	docker-compose -f ./localfiles/uptrace.compose.yaml down
