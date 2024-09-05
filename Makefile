all: migration clean server

migration:
	@echo "Running Migrations..."
	cp ./config/local.yml ./config.yml
	docker compose up -d
	go build -o elasticsearch-migrations ./cmd/app/main.go 
	./elasticsearch-migrations -migrate

clean:
	rm -rf ./elasticsearch-migrations

server:
	go build -o service_catalogue ./cmd/app/main.go
	./service_catalogue
