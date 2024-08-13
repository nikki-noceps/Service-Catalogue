server:
	@echo "Running Service Catalogue http server..."
	cp ./config-files/local.yml ./config.yml
	go build -o service_catalogue ./main.go
	./service_catalogue
