include ../../docker.mk

#DOCKER         ?= docker
#DOCKER_COMPOSE ?= docker-compose
#DOCKER_REPO    ?= ""
#DOCKER_TAG     ?= latest
#DOCKER_BUILD   ?= $(DOCKER) build --rm --no-cache --pull

MAP_TAG?=update-max-code-size
NETWORK_ID?=214

# To check shard endpoints and restart Harmony network if necessary
RETRY ?= 5

SCRIPT_DIR=./scripts
CONTRACT_DIR=../../../contract
OUTPUT_ADDRESS_DIR=./output/addresses

.PHONY: docker-image
docker-image:
	make up-scaffold-with-check
	make deploy-contract
	make save-contract-address
	make docker-commit
	make down-scaffold

.PHONY: deploy-contract
deploy-contract:
	cd ${CONTRACT_DIR} && npm install && npm run migrate

.PHONY: up-scaffold-with-check
# This starts harmony local network with verifying that
# each explorer node exists in the shard specified in config.
# If the verification fails, it restarts the network.
up-scaffold-with-check:
	make up-scaffold ; \
	make wait-for-launch ATTEMPT=30 CONTAINER=map-scaffold ; \

.PHONY: up-scaffold
up-scaffold:
	${DOCKER_COMPOSE} up --build -d

.PHONY: down-scaffold
down-scaffold:
	${DOCKER_COMPOSE} down --volumes --remove-orphans

.PHONY: save-contract-address
save-contract-address:
	mkdir -p ${OUTPUT_ADDRESS_DIR}
	${SCRIPT_DIR}/saveContractAddresses.sh ${NETWORK_ID} ${CONTRACT_DIR} ${OUTPUT_ADDRESS_DIR}

.PHONY: docker-commit
docker-commit:
	${SCRIPT_DIR}/commitImage.sh  ${DOCKER_REPO} ${DOCKER_TAG} map-chain map-scaffold ${OUTPUT_ADDRESS_DIR}
