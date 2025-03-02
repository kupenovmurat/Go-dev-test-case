# Примеры использования API

## Загрузка файла

### Запрос

```bash
curl -X POST -F "file=@/path/to/your/file.txt" http://localhost:8080/upload
```

### Ответ

```json
{
  "fileId": "550e8400-e29b-41d4-a716-446655440000",
  "name": "file.txt",
  "size": 1024
}
```

## Скачивание файла

### Запрос

```bash
curl -X GET http://localhost:8080/download/550e8400-e29b-41d4-a716-446655440000 -o downloaded_file.txt
```

## Получение списка файлов

### Запрос

```bash
curl -X GET http://localhost:8080/files
```

### Ответ

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "file.txt",
    "size": 1024,
    "contentType": "text/plain",
    "createdAt": "2023-09-15T12:00:00Z",
    "chunkCount": 6
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "image.jpg",
    "size": 2048,
    "contentType": "image/jpeg",
    "createdAt": "2023-09-15T12:30:00Z",
    "chunkCount": 6
  }
]
```

## Удаление файла

### Запрос

```bash
curl -X DELETE http://localhost:8080/files/550e8400-e29b-41d4-a716-446655440000
```

### Ответ

```json
{
  "success": true
}
```

## Регистрация сервера хранения

### Запрос

```bash
curl -X POST -H "Content-Type: application/json" -d '{"url": "http://localhost:8081"}' http://localhost:8080/storage/register
```

### Ответ

```json
{
  "id": "storage-1",
  "url": "http://localhost:8081",
  "available": true,
  "usedSpace": 0
}
```

## Получение списка серверов хранения

### Запрос

```bash
curl -X GET http://localhost:8080/storage/servers
```

### Ответ

```json
[
  {
    "id": "storage-1",
    "url": "http://localhost:8081",
    "available": true,
    "usedSpace": 1024
  },
  {
    "id": "storage-2",
    "url": "http://localhost:8082",
    "available": true,
    "usedSpace": 2048
  }
]
``` 