# GOBrokers

## REDIS

### Кеширование

```bash
sudo docker-compose up -d
cd redis-cache
go run main.go
```

#### Чистим кеш

```bash
redis-cli
FLUSHDB
KEYS "profile:*"
exit
```

#### Проверяем, что кеш работает

```bash
curl -i http://localhost:8080/profile/123
```

- X-Cache: MISS = НЕ КЕШ
- X-Cache: HIT = КЕШ


Парсинг кеша/не кеша
```bash
curl -s -o /dev/null -D - http://localhost:8080/profile/123 | grep -i X-Cache
```

### PubSub

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

## RABBITMQ-TASK

### Реализовать отправку задач в очередь и обработку их воркером (например, задача на отправку email).

На одном терминале
```bash
sudo docker-compose up -d
cd rabbitmq-task
go run main.go publisher.go consumer.go nats_example.go consumer
```

На другом терминале
```bash
cd rabbitmq-task
go run main.go publisher.go consumer.go nats_example.go publisher
```

Можем также через UI посмотреть состояние очереди. (логин: guest, пароль: guest)
```bash
curl http://localhost:15672/
```

### NATS

На одном терминале
```bash
sudo docker-compose up -d
cd rabbitmq-task
go run main.go publisher.go consumer.go nats_example.go nats
```

## CQRS

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
