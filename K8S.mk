.PHONY: configure-sample upload-code do-upload-code upload-app

configure-sample: dist build
	kubectl create ns lsd --dry-run=client --save-config -o yaml \
		| kubectl apply -f -
	./dist/lsd gen k8s -pull-if-not-present \
		-s landing-page:frontend \
		-s api:backend \
		-s bff:backend \
		-s mysql:database \
		-s redis:cache \
		-i landing-page:9000:landing.localhost:/ \
		-i manager:9001:manager.localhost:/ \
			| kubectl apply --namespace lsd -f -

upload-code: | dist do-upload-code

endpoint?=http://manager.localhost
serviceType?=backend
codefile?=./examples/$(serviceType)/handler.lua
do-upload-code:
	cat $(codefile) \
		| ./dist/lsd ctl manager --endpoint $(endpoint) \
			upload --service-type $(serviceType)

upload-app: dist
	make -C . do-upload-code serviceType=backend
	make -C . do-upload-code serviceType=frontend
	make -C . do-upload-code serviceType=database
