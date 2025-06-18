run:
	@trap 'echo "Process interrupted, exiting..."; exit 0' INT; \
	go run cmd/server/main.go

migrate_create:
	migrate create -ext=sql -dir=migrations -seq init

migrate_up:
	migrate -path=migrations -database "mysql://user:password@tcp(localhost:3306)/travel_db" up

migrate_down:
	migrate -path=migrations -database "mysql://user:password@tcp(localhost:3306)/travel_db" down

migrate_version:
	migrate -path=migrations -database "mysql://user:password@tcp(localhost:3306)/travel_db" version

migrate_force:
	migrate -path=migrations -database "mysql://user:password@tcp(localhost:3306)/travel_db" force 1

sqlc:
	sqlc generate

compose_up:
	docker compose up -d

compose_down:
	docker compose down

compose_logs:
	docker compose logs -f

.PHONY: run migrate_create migrate_up migrate_down migrate_version migrate_force sqlc compose_up compose_down compose_logs
