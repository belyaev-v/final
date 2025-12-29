# Инструкция по запуску проекта

## Настройка базы данных

Можно запустить start_db в Makefile

Если у вас другие параметры подключения, используйте флаги `-dsn` при запуске сервисов.


## Порты сервисов по дефолту

- **APIGateway**: 8080
- **NewsService**: 8081
- **CommentsService**: 8082
- **CensorshipService**: 8083

## Тестирование через API Gateway

Все запросы делаются через API Gateway на порту 8080.

### Примеры запросов:

1. **Получить список новостей:**
```bash
curl http://localhost:8080/news
```

2. **Получить список новостей с пагинацией:**
```bash
curl http://localhost:8080/news?page=1
```

3. **Поиск новостей:**
```bash
curl http://localhost:8080/news?s=Go
```

4. **Получить новость с комментариями:**
```bash
curl http://localhost:8080/news/1
```

5. **Создать комментарий (валидный):**
```bash
curl -X POST http://localhost:8080/news/1/comments \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-123" \
  -d '{"text": "Отличная новость!"}'
```

6. **Создать комментарий (невалидный - содержит запрещенное слово):**
```bash
curl -X POST http://localhost:8080/news/1/comments \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-456" \
  -d '{"text": "Это qwerty комментарий"}'
```

## Проверка логов

Все сервисы логируют HTTP-запросы в stdout

Request ID передается через заголовок `X-Request-ID` и проходит через всю цепочку сервисов.

## Остановка сервисов

Все сервисы корректно обрабатывают сигналы SIGTERM и SIGINT для graceful shutdown.

