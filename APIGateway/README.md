
```bash
go run main.go -port=8080 -news-url=http://localhost:8081 -comments-url=http://localhost:8082 -censorship-url=http://localhost:8083
```

## Эндпоинты

- `GET /news` - список новостей (поддерживает параметры `?s=keyword` для поиска и `?page=N` для пагинации)
- `GET /news/filter` - фильтр новостей (аналогично `/news`)
- `GET /news/{id}` - детальная новость с комментариями
- `POST /news/{id}/comments` - создание комментария к новости
