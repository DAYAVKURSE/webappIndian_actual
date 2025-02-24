.PHONY: postgres api api-test api-start-with-test tgbot site migrate clear

postgres:
	docker compose -f backend/postgres/compose.yml up -d

api:
	cd backend/api && docker compose -f compose.yml up -d api redis --build

api-test:
	cd backend/api && docker compose -f compose.yml up --exit-code-from api_test api_test

api-start-with-test:
	@cd backend/api && docker compose -f compose.yml up --exit-code-from api_test api_test
	@if [ $$? -eq 0 ]; then \
		echo "Tests passed. Starting API..."; \
		cd backend/api && docker compose -f compose.yml up -d api redis; \
	else \
		echo "Tests failed. API will not be started."; \
		exit 1; \
	fi

tgbot:
	cd backend/telegramBot && docker compose -f compose.yml up -d --build

site:
	cd site && docker compose -f compose.yml up --build -d

migrate:
	cd backend/api && \
	docker build -f Dockerfile.migrate -t api-migrate . && \
	docker run --rm \
		--network db-network \
		--env-file ../.env \
		-v /var/log/BlessedLogs/:/logs/ \
		api-migrate

clear:
	@echo "Stopping and removing all containers..."
	@docker stop $$(docker ps -aq) 2>/dev/null || true
	@docker rm $$(docker ps -aq) 2>/dev/null || true
	@echo "Removing all images..."
	@docker rmi $$(docker images -q) 2>/dev/null || true
	@echo "Pruning the system..."
	@docker system prune -af
	@echo "Cleanup complete."1

