init:
	docker-compose down
	docker-compose build
	docker-compose up -d
	docker-compose ps
