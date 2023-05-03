postgres:
	docker run --network bank-network --name simple-bank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -d postgres
createdb:
	docker exec -it simple-bank createdb --username=root --owner=root simple-bank
dropdb:
	docker exec -it simple-bank dropdb simple-bank
migrateup:
	migrate -path ./db/migration -database "postgresql://root:dbrTIf%3Cfx%23l9%3FKYi9JZ6%24Q%5BvJH1%26@simple-bank.c7ihrxlceufb.ap-southeast-1.rds.amazonaws.com:5432/simple_bank" -verbose up
migrateup1:
	migrate -path ./db/migration -database "postgresql://root:dbrTIf%3Cfx%23l9%3FKYi9JZ6%24Q%5BvJH1%26@simple-bank.c7ihrxlceufb.ap-southeast-1.rds.amazonaws.com:5432/simple_bank" -verbose up 1
migratedown:
	migrate -path ./db/migration -database "postgresql://root:dbrTIf%3Cfx%23l9%3FKYi9JZ6%24Q%5BvJH1%26@simple-bank.c7ihrxlceufb.ap-southeast-1.rds.amazonaws.com:5432/simple_bank" -verbose down
migratedown1:
	migrate -path ./db/migration -database "postgresql://root:dbrTIf%3Cfx%23l9%3FKYi9JZ6%24Q%5BvJH1%26@simple-bank.c7ihrxlceufb.ap-southeast-1.rds.amazonaws.com:5432/simple_bank" -verbose down 1
sqlc:
	docker run --rm -v "D:\Code\simple-bank:/src" -w /src kjconroy/sqlc:1.17.2 generate
test:
	go test -v --cover ./...
server:
	go run main.go
mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/patchbrain/simple-bank/db/sqlc Store
.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock