# Распределенная система хранения файлов

Это мой тестовый проект распределенной системы хранения файлов, вдохновленный Amazon S3, но с некоторыми уникальными особенностями.

## Требования

Для запуска проекта вам потребуется:
- Go 1.21 или выше
- Docker и Docker Compose (для запуска через контейнеры)
- Минимум 512 МБ свободной оперативной памяти
- Около 100 МБ свободного места на диске

## Как это работает

Система состоит из двух основных компонентов:
- **REST-сервер** - принимает файлы от пользователей, разбивает их на 6 частей и распределяет по серверам хранения
- **Серверы хранения** - хранят части файлов и отдают их по запросу

## Особенности

- Поддержка файлов размером до 10 ГБ
- Равномерное распределение нагрузки между серверами хранения
- Возможность динамически добавлять новые серверы хранения
- Устойчивость к обрывам соединения при загрузке
- Простая настройка через Docker Compose

## Запуск

Самый простой способ запустить систему - использовать Docker Compose:

```bash
docker-compose up
```

Или можно запустить компоненты по отдельности:

```bash
# Запуск REST-сервера
go run ./cmd/rest-server/main.go

# Запуск сервера хранения
go run ./cmd/storage-server/main.go --port=8081 --id=storage-1
```

## API

- `POST /upload` - загрузка файла
- `GET /download/:fileId` - скачивание файла
- `GET /files` - список файлов
- `DELETE /files/:fileId` - удаление файла

## Тестирование

Для проверки работоспособности системы можно использовать тестовый клиент:

```bash
go run ./cmd/test-client/main.go
``` 