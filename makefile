postgres:
	docker run --name simple-bank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres
createdb:
	docker exec -it simple-bank createdb --username=root --owner=root simple-bank
dropdb:
	docker exec -it simple-bank dropdb simple-bank
migrateup:
	migrate -path ./db/migration -database "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable" -verbose up
migratedown:
	migrate -path ./db/migration -database "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable" -verbose down
sqlc:
	docker run --rm -v "D:\Code\simple-bank:/src" -w /src kjconroy/sqlc generate
test:
	go test -v --cover ./...
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test