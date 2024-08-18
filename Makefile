server:
	@echo "Running Service Catalogue http server..."
	cp ./config/local.yml ./config.yml
	docker compose up -d
	go build -o service_catalogue ./main.go
	./service_catalogue
