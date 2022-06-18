IMAGE_REPO?=andrebq/lsd
IMAGE_TAG?=latest
IMAGE_FULL_NAME=$(IMAGE_REPO):$(IMAGE_TAG)

CONTAINER_NAME?=learn-system-design
LOCAL_PORT?=9000
STRESS_LOCAL_PORT?=9001
CONTROL_PLANE_LOCAL_PORT?=9002

LSD_CONTROL_PLANE_BIND?=127.0.0.1:$(CONTROL_PLANE_LOCAL_PORT)
LSD_STRESS_TEST_SERVE?=127.0.0.1:$(STRESS_LOCAL_PORT)
