.PHONY: all build test clean sqlc-generate tidy lint

# Variables
BINARY_DIR=bin
GO=go

all: build test

## sqlc-generate: Genera el código tipado para la capa de persistencia a partir de SQL
sqlc-generate:
	@echo "==> Generando código Go con sqlc..."
	sqlc generate

## test: Ejecuta los tests unitarios
test:
	@echo "==> Ejecutando tests unitarios..."
	$(GO) test -v ./...

## test-integration: Ejecuta los tests de integración (requiere build tag 'integration')
test-integration:
	@echo "==> Ejecutando tests de integración..."
	$(GO) test -v -tags=integration ./test/integration/...

## docker-up: Levanta PostgreSQL 16 y Redis 7 para desarrollo e integración
docker-up:
	@echo "==> Levantando servicios de infraestructura con docker-compose..."
	docker-compose -f deployments/docker-compose.yml up -d

## docker-down: Detiene y elimina contenedores de infraestructura
docker-down:
	@echo "==> Deteniendo contenedores..."
	docker-compose -f deployments/docker-compose.yml down -v

## test-race: Ejecuta los tests unitarios con detección de condiciones de carrera
test-race:
	@echo "==> Ejecutando tests con race detector..."
	$(GO) test -v -race ./...

## tidy: Descarga y limpia módulos de Go
tidy:
	@echo "==> Ejecutando go mod tidy..."
	$(GO) mod tidy

## build: Compila todos los paquetes del monorepo y genera los binarios
build:
	@echo "==> Compilando paquetes..."
	$(GO) build ./...
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/collector ./cmd/collector
	$(GO) build -o $(BINARY_DIR)/dispatcher ./cmd/dispatcher
	$(GO) build -o $(BINARY_DIR)/api-server ./cmd/api-server

## docs-pdf: Genera el documento PDF de arquitectura y especificación técnica
docs-pdf:
	@echo "==> Generando documentación técnica en PDF..."
	python3 docs/generate_pdf.py

## clean: Limpia artefactos generados y temporales
clean:
	@echo "==> Limpiando binarios y temporales..."
	rm -rf $(BINARY_DIR)

