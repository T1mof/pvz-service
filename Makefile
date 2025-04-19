.PHONY: dev build test clean logs restart db-shell db-migrate

# Переменные
DOCKER_COMPOSE = docker-compose
GO = go

# Основные команды
dev:
	$(DOCKER_COMPOSE) up -d
	@echo "Среда разработки запущена. Доступно по адресу http://localhost:8080"

build:
	$(DOCKER_COMPOSE) build

test:
	$(GO) test ./...

clean:
	$(DOCKER_COMPOSE) down -v
	rm -rf tmp

# Команды для работы с базой данных
db-migrate:
	$(DOCKER_COMPOSE) exec app /app/scripts/init-db.sh

db-reset:
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d
	@echo "База данных сброшена, ожидаем инициализации..."
	sleep 10
	@echo "База данных готова"

# Вспомогательные команды
logs:
	$(DOCKER_COMPOSE) logs -f

restart:
	$(DOCKER_COMPOSE) restart app

db-shell:
	$(DOCKER_COMPOSE) exec db psql -U postgres -d pvz

service-status:
	$(DOCKER_COMPOSE) ps