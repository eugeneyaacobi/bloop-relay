GO ?= go
DOCKER_COMPOSE ?= docker compose
DOCKER_TEST_COMPOSE := deploy/compose/docker-test.yml

.PHONY: fmt test build relay client smoke-docker verify

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	$(GO) build ./...

relay:
	$(GO) run ./cmd/bloop-relay

client:
	$(GO) run ./cmd/bloop-client

smoke-docker:
	@set -euo pipefail; \
	trap '$(DOCKER_COMPOSE) -f $(DOCKER_TEST_COMPOSE) down' EXIT; \
	$(DOCKER_COMPOSE) -f $(DOCKER_TEST_COMPOSE) up --build -d; \
	sleep 2; \
	curl -si --fail -H 'Host: public.bloop.to' http://127.0.0.1:38080/hello; \
	printf '\n---\n'; \
	curl -si --fail -u gene:secretpass -H 'Host: basic.bloop.to' http://127.0.0.1:38080/hello; \
	printf '\n---\n'; \
	curl -si --fail -H 'Host: token.bloop.to' -H 'X-Bloop-Token: topsecret' http://127.0.0.1:38080/hello; \
	printf '\n---\n'; \
	curl -si --fail -X POST -H 'Host: public.bloop.to' --data 'ping=post-body' http://127.0.0.1:38080/submit

verify: fmt test smoke-docker
