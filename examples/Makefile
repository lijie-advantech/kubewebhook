

POD_ANNOTATE_IMAGE="harbor.wise-paas.io/li.jie/kubewebhook-pod-annotate-example"
POD_ANNOTATE_EXAMPLE="pod-annotate"
INGRESS_HOST_VALIDATOR_IMAGE="harbor.wise-paas.io/li.jie/kubewebhook-ingress-host-validator-example"
INGRESS_HOST_VALIDATOR_EXAMPLE="ingress-host-validator"

# Build the development docker image
.PHONY: build-examples
build-examples:
	docker build -t $(POD_ANNOTATE_IMAGE) --build-arg example=$(POD_ANNOTATE_EXAMPLE) -f ./Dockerfile ../
#	docker build -t $(INGRESS_HOST_VALIDATOR_IMAGE) --build-arg example=$(INGRESS_HOST_VALIDATOR_EXAMPLE) -f ./Dockerfile ../

.PHONY: push-examples
push-examples:
	docker push $(POD_ANNOTATE_IMAGE)
#	docker push $(INGRESS_HOST_VALIDATOR_IMAGE)

.PHONY: create-certs
create-certs:
	./create-certs.sh default ${POD_ANNOTATE_EXAMPLE}
	./create-certs.sh default ${INGRESS_HOST_VALIDATOR_EXAMPLE}
