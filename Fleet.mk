.PHONY: start-fleet

start-fleet: dist
	cp -Rv ./examples/* ./dist/examples/
	./dist/lsd --otel.service.name "fleet-manager" \
		fleet serve-local \
		--scriptBase $(PWD)/examples/ \
		--app echo-handler \
		--app backend \
		--app frontend \
		--app database \
