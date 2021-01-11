# https://myapollo.com.tw/zh-tw/docker-mysql/
server_name = "sandy230207/mask-server:v2"
db_name = "sandy230207/mask-db:v1"

db-build:
	docker build \
		-f db.Dockerfile \
		-t $(db_name) .

db-run:
	docker run -d \
		--name=mask-db \
		-p 3306:3306 \
		--env MYSQL_ROOT_PASSWORD=password \
		$(db_name)

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
		-t $(server_name) .

run:
	docker run -d \
		--name=mask-server \
		-p 3000:3000 \
		$(server_name)

server-push:
	docker push $(server_name)

db-push:
	docker push $(db_name)

# docker run --name=mask-db -p 3306:3306 --env MYSQL_ROOT_PASSWORD=password "sandy230207/mask-db:v1"
# docker run --name=mask-server -p 3000:3000 "sandy230207/mask-server:v1"
