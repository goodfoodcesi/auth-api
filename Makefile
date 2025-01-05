build:
	docker build .
run:
	docker compose -f docker-compose-rabbitmq.yaml -p rabbitmq up  -d
	docker compose up -d --build
	docker compose -f docker-compose-traefik.yaml -p traefik  up -d
stop:
	docker compose down
logs:
	docker logs -f auth-api-authapi-1
migrate:
	migrate -path db/migration -database  "postgresql://authapi:authapi@localhost:5432/authapi?sslmode=disable" -verbose up
migrate-down:
	migrate -path db/migration -database  "postgresql://authapi:authapi@localhost:5432/authapi?sslmode=disable" -verbose down
