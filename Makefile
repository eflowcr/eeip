.PHONY: build test docker-build docker-push deploy logs restart stop migrate clean

build:
	cd backend && go build -o bin/api cmd/api/main.go
	cd frontend && npm run build

test:
	cd backend && go test ./...

docker-build:
	docker build -t ${DOCKER_USERNAME}/eeip-backend:latest ./backend
	docker build -t ${DOCKER_USERNAME}/eeip-frontend:latest ./frontend

docker-push:
	docker push ${DOCKER_USERNAME}/eeip-backend:latest
	docker push ${DOCKER_USERNAME}/eeip-frontend:latest

deploy:
	docker compose -f docker-compose.prod.yml pull
	docker compose -f docker-compose.prod.yml up -d

logs:
	docker compose logs -f

restart:
	docker compose restart

stop:
	docker compose down

migrate:
	docker compose exec backend /app/bin/api -migrate-only

clean:
	docker compose down -v
	rm -rf backend/bin
