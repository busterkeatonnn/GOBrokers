# GOBrokers

- PR7 - REDIS (CACHE + PUBSUB) + (RABBITMQ-TASK + NATS) + CQRS
- PR11 - REDIS(PUBSUB)
- PR12 - RABBITMQ-TASK (Использовать manual ack, В случае ошибки — не подтверждать сообщение)
- PR13 - NATS - сделать несколько Consumer'ов, которые подключатся к NATS, подпишутся на тему jobs.create через queue group workers, а потом каждый обработчик будет обрабатывать только свою часть задач.


## PR7
### REDIS

#### Кеширование

```bash
sudo docker-compose up -d
cd redis-cache
go run main.go
```

##### Чистим кеш

```bash
redis-cli
FLUSHDB
KEYS "profile:*"
exit
```

##### Проверяем, что кеш работает

```bash
curl -i http://localhost:8080/profile/123
```

- X-Cache: MISS = НЕ КЕШ
- X-Cache: HIT = КЕШ


Парсинг кеша/не кеша
```bash
curl -s -o /dev/null -D - http://localhost:8080/profile/123 | grep -i X-Cache
```

---

### RABBITMQ-TASK

#### Реализовать отправку задач в очередь и обработку их воркером (например, задача на отправку email).

На одном терминале
```bash
sudo docker-compose up -d
cd rabbitmq-task
go run *.go consumer
```

На другом терминале
```bash
cd rabbitmq-task
go run *.go publisher
```

Можем также через UI посмотреть состояние очереди. (логин: guest, пароль: guest)
```bash
curl http://localhost:15672/
```

#### NATS

Для PR7 На одном терминале
```bash
sudo docker-compose up -d
cd rabbitmq-task
go run *.go nats
```

Для PR13 На одном терминале
```bash
sudo docker-compose up -d
cd rabbitmq-task
go run *.go nats-publisher
```

Для PR13 На других терминалах
```bash
cd rabbitmq-task
# Терминал 1
WORKER_ID=worker1 go run *.go nats-consumer

# Терминал 2
WORKER_ID=worker2 go run *.go nats-consumer

# Терминал 3
WORKER_ID=worker3 go run *.go nats-consumer
```
---

### CQRS

```bash
sudo docker-compose up -d
cd cqrs-example
go run *.go
```

Создайте заказ:
```bash
curl -X POST "http://localhost:8081/orders?customer_id=user123&item=книга"
```

Получите список заказов:
```bash
curl http://localhost:8081/orders
```

Получите информацию о конкретном заказе (замените 1 на ID вашего заказа):
```bash
curl http://localhost:8081/orders/1
```

Оплатите заказ:
```bash
curl -X POST http://localhost:8081/orders/1/pay
```

Отмените заказ:
```bash
curl -X POST "http://localhost:8081/orders/1/cancel?reason=Передумал"
```

---

## PR11
### REDIS (PubSub)

На одном терминале:
```bash
sudo docker-compose up -d
cd redis-cache
go run pubsub.go subscriber
```

На другом терминале:
```bash
cd redis-cache
go run pubsub.go subscriber
```
