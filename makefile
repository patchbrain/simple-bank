postgres:
	docker run --name my-postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres
createdb:
	docker exec -it my-postgres createdb --username=root --owner=root simple-bank
dropdb:
	docker exec -it my-postgres dropdb simple-bank
migrateup:
	migrate -path ./db/migration -database "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable" -verbose up
migratedown:
	migrate -path ./db/migration -database "postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable" -verbose down
sqlc:
	docker run --rm -v "D:\Work\GolangMasterClass\Simple-bank:/src" -w /src kjconroy/sqlc:1.17.2 generate
test:
	go test -v --cover ./...
server:
	go run main.go
mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/patchbrain/simple-bank/db/sqlc Store
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock