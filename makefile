DB_URL=postgresql://root:123456@localhost:5432/simple-bank?sslmode=disable

postgres:
	docker run --network bank-network --name simple-bank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres
createdb:
	docker exec -it simple-bank createdb --username=root --owner=root simple-bank
dropdb:
	docker exec -it simple-bank dropdb simple-bank
migrateup:
	migrate -path ./db/migration -database "$(DB_URL)" -verbose up
migrateup1:
	migrate -path ./db/migration -database "$(DB_URL)" -verbose up 1
migratedown:
	migrate -path ./db/migration -database "$(DB_URL)" -verbose down
migratedown1:
	migrate -path ./db/migration -database "$(DB_URL)" -verbose down 1
sqlc:
	docker run --rm -v "D:\Code\simple-bank:/src" -w /src kjconroy/sqlc:1.17.2 generate
test:
	go test -v --cover ./...
server:
	go run main.go
db_docs:
	dbdocs build doc/db.dbml
db_schema:
	dbml2sql --postgres -o doc/scheme.sql doc/db.dbml
mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/patchbrain/simple-bank/db/sqlc Store
proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative proto/*.proto
evans:
	evans --port 8081 --host localhost -r repl
.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock db_docs db_schema proto evans