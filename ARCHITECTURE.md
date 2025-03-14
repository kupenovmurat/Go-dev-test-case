# Архитектура системы

## Общий обзор

Система состоит из следующих компонентов:

1. **REST-сервер (Server A)** - центральный компонент, который обрабатывает запросы клиентов
2. **Серверы хранения (Servers Bn)** - серверы для хранения частей файлов
3. **Тестовый модуль** - для проверки функциональности системы

## REST-сервер

REST-сервер выполняет следующие функции:
- Принимает файлы от клиентов
- Разбивает файлы на 6 равных частей
- Распределяет части по серверам хранения
- Хранит метаданные о файлах и их частях
- Собирает части файлов при запросе на скачивание

## Серверы хранения

Серверы хранения выполняют следующие функции:
- Хранят части файлов
- Предоставляют API для загрузки и скачивания частей

## Процесс загрузки файла

1. Клиент отправляет файл на REST-сервер
2. REST-сервер проверяет наличие достаточного количества серверов хранения
3. REST-сервер разбивает файл на 6 равных частей
4. Каждая часть загружается на отдельный сервер хранения
5. REST-сервер сохраняет метаданные о файле и его частях

## Процесс скачивания файла

1. Клиент запрашивает файл по его ID
2. REST-сервер получает метаданные файла
3. REST-сервер скачивает все части файла с серверов хранения
4. REST-сервер объединяет части в исходный файл
5. REST-сервер отправляет файл клиенту

## Обработка ошибок

- Если сервер хранения недоступен при загрузке, процесс загрузки прерывается
- Если сервер хранения недоступен при скачивании, клиенту возвращается ошибка
- Если клиент отключается во время загрузки, загрузка прерывается и временные данные удаляются 