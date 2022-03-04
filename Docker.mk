.PHONY: build run stop

build:
	docker build -t $(IMAGE_FULL_NAME) -f Dockerfile .

run:
	docker run --rm -ti \
		-v $(PWD):$(PWD) \
		-w $(PWD) \
		-p $(LOCAL_PORT):8080 \
		--name $(CONTAINER_NAME) \
		$(IMAGE_FULL_NAME)

stop:
	docker kill $(CONTAINER_NAME)
