
```bash
go run main.go -port=8082 -dsn="postgres://user:password@localhost/commentsdb?sslmode=disable"
```

## Эндпоинты

- `POST /comments` - создание комментария
  - Body: `{"news_id": 1, "text": "Комментарий", "parent_comment_id": null}`
- `GET /comments?news_id={id}` - получение всех комментариев по новости

