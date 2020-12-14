# https://myapollo.com.tw/zh-tw/docker-mysql/
db-build:
	docker build \
		-f db.Dockerfile \
		-t mask/db:v1 .

db-run:
	docker run \
		--name=mask-db \
		-p 3306:3306 \
		--env MYSQL_ROOT_PASSWORD=password \
		mask/db:v1

db-run2: 
	docker run -d \
		--name=mysql \
		-p 3306:3306 \
		--env MYSQL_ROOT_PASSWORD=password \
		mysql/mysql-server

db-login2:
	docker exec -it mysql mysql -u root -p

build:
	docker build \
		-f server.Dockerfile \
		-t mask/server:v1 .

run:
	docker run -d \
		--name=mask-server \
		-p 3000:3000 \
		mask/server:v1
