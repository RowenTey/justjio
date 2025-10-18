DB_CONTAINER=postgres
BACKEND_CONTAINER=backend

# Start containers
up:
	cd infra/docker && docker compose up -d

# Stop containers
down:
	cd infra/docker && docker compose down -v

# Restart containers
restart:
	cd infra/docker && docker compose down -v && docker compose up -d
	
# Start containers with observability tools
lgtm-up:
	cd infra/docker && docker compose -f docker-compose.yml -f docker-compose-tempo.yml up -d
	
# Stop containers with observability tools
lgtm-down:
	cd infra/docker && docker compose -f docker-compose.yml -f docker-compose-tempo.yml up -d

# Rebuild everything and start containers
rebuild:
	cd infra/docker && docker compose up --build

# Show logs
logs:
	cd infra/docker && docker compose logs -f

# Open a shell inside backend container
shell:
	cd infra/docker && docker compose exec $(BACKEND_CONTAINER) /bin/sh

# Run the database shell
dbshell:
	cd infra/docker && docker compose exec -it $(DB_CONTAINER) psql -U postgres

# Reset database
reset-db:
	cd infra/docker && docker compose exec $(DB_CONTAINER) psql -U postgres -d postgres -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	
# Generate API documentation
swagger:
	cd server/api && swag init -g main.go
	
# Start API server 
api:
	cd server/api && air dev