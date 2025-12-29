
```bash
go run main.go -port=8081 -dsn="postgres://user:password@localhost/newsdb?sslmode=disable"
```

## Эндпоинты

- `GET /news` - список новостей с пагинацией и поиском
  - Параметры: `?page=N` (номер страницы), `?s=keyword` (поиск по заголовку)
- `GET /news/{id}` - детальная информация о новости
