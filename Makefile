# develop-cmd section.
build-develop:
docker compose -f deployments/develop/docker-compose-develop.yml build

run-develop:
docker compose -f deployments/develop/docker-compose-develop.yml up
