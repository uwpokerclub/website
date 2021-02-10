.PHONY: help
PYTHON=python

help:   ## Show this help.
	$(info Available targets)
	@awk '/^[a-zA-Z\-\_0-9]+:/ {                                   \
          nb = sub( /^## /, "", helpMsg );                             \
          if(nb == 0) {                                                \
            helpMsg = $$0;                                             \
            nb = sub( /^[^:]*:.* ## /, "", helpMsg );                  \
          }                                                            \
          if (nb)                                                      \
            printf "\033[1;31m%-" width "s\033[0m %s\n", $$1, helpMsg; \
        }                                                              \
        { helpMsg = $$0 }'                                             \
        width=$$(grep -o '^[a-zA-Z_0-9]\+:' $(MAKEFILE_LIST) | wc -L)  \
	$(MAKEFILE_LIST)

dev-rebuild:  ## tear down and rebuild dev environment
	docker-compose down
	docker-compose build
	docker-compose up -d
	docker exec -ti uwpokerclub_app_app_1 sh

dev:  ## log into to dev env. create one if there is none
	docker-compose up -d
	docker exec -ti uwpokerclub_app_app_1 sh

shell:
  docker exec -ti uwpokerclub_app_app_1 sh

down:
  docker-compose down
