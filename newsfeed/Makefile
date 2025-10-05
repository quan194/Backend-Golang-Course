# variable
MYSQL_SVC=newsfeed-mysql
REDIS_SVC=newsfeed-redis
KAFKA_SVC=newsfeed-kafka

MYSQL_USERNAME=user
MYSQL_PASSWORD=pass1234
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=newsfeed
MYSQL_CONN=mysql://$(MYSQL_USERNAME):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)
MYSQL_MIGRATION_FOLDER=script/db_migration

REDIS_HOST=localhost
REDIS_PORT=6379

KAFKA_HOST=localhost
KAFKA_PORT=9092
KAFKA_TOPIC=posts
KAFKA_CONSUMER_GROUP=newsfeed_worker

# docker: for local dev
up: # up all services
	docker-compose up -d
down: # down all services
	docker-compose down

exec_db: # access mysql container
	docker exec -it $(MYSQL_SVC) mysql -u$(MYSQL_USERNAME) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)

exec_redis: # access redis container
	docker exec -it $(REDIS_SVC) redis-cli

# database migration
STEP ?=1
migrate_up_db: # run migrate sql (by step number, default 1)
	migrate -path $(MYSQL_MIGRATION_FOLDER) -database "$(MYSQL_CONN)" up $(STEP)

migrate_down_db: # run downgrade sql (by step number, default 1)
	migrate -path $(MYSQL_MIGRATION_FOLDER) -database "$(MYSQL_CONN)" down $(STEP)

migrate_status: # check current version
	migrate -path $(MYSQL_MIGRATION_FOLDER) -database "$(MYSQL_CONN)" version

# kafka
kafka_init_topic: # only run if topic is not created, run after container is up
	docker exec -it $(KAFKA_SVC) kafka-topics.sh \
      --create \
      --topic $(KAFKA_TOPIC) \
      --bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT) \
      --partitions 3 \
      --replication-factor 1

kafka_topic_status:
	# list all topics
	docker exec -it $(KAFKA_SVC) kafka-topics.sh --list --bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT)
	# describe topic `posts`
	docker exec -it $(KAFKA_SVC) kafka-topics.sh \
		--describe \
		--topic $(KAFKA_TOPIC) \
		--bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT)

kafka_consumer_status:
    # list consumer groups
	docker exec -it $(KAFKA_SVC) kafka-consumer-groups.sh \
       --bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT) \
       --list
   # describe consumer group `newsfeed-worker`
	docker exec -it $(KAFKA_SVC) kafka-consumer-groups.sh \
		--bootstrap-server $(KAFKA_HOST):$(KAFKA_PORT) \
		--describe \
		--group $(KAFKA_CONSUMER_GROUP)

# gen proto
gen_proto:
	protoc \
      --go_out=. --go_opt=paths=source_relative \
      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
      internal/handler/proto/*/*.proto
