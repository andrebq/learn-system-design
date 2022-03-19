.PHONY: jaeger-up jaeger-down


jaeger-up: dist
	cd localfiles && ../dist/lsd support jaeger docker-compose
	docker-compose -f ./localfiles/jaeger.compose.yaml up

jaeger-down:
	docker-compose -f ./localfiles/jaeger.compose.yaml down

grafana-up: dist
	cd localfiles && ../dist/lsd support grafana docker-compose -compose-file grafana.compose.yaml
	docker-compose -f ./localfiles/grafana.compose.yaml up

grafana-down:
	docker-compose -f ./localfiles/grafana.compose.yaml down
