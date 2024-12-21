# calculator-service

## Описание проекта

Этот сервис предоставляет веб-API для вычисления арифметических выражений. Пользователь отправляет HTTP запрос с
выражением, а в ответ получает результат вычисления или сообщение об ошибке.

Сценарии работы сервиса:

- Успешное вычисление выражения
- Ошибка валидации выражения
- Внутренняя ошибка сервера

---

## Требования

Go **1.19+**

---

## Запуск

1. Склонировать репозиторий
   ```bash
   git clone https://github.com/alexGoLyceum/calculator-service.git
   ```
2. Перейти в папку с проектом
   ```bash
   cd calculator-service
   ```
3. Проверить, что все зависимости установлены
   ```bash
   go mod tidy
   ```
4. Запустить сервис
   ```bash
   go run ./cmd/server/main.go
   ```
5. Сервер будет запущен на `localhost:8080`. Если надо поменять хост и порт, то можно сделать это в файле `config.yaml`

## Возможные ответы сервиса

1. Успешное вычисление
    ```json
    {
      "result": "5"
    }
    ```
2. Неверное выражение
    ```json
    {
      "error": "Expression is not valid"
    }
    ```

3. Внутренняя ошибка сервиса
    ```json
    {
      "error": "Internal server error"
    }
    ```

---

## Эндпоинты

`POST /api/v1/calculate`

Сервис принимает JSON-объект с арифметическим выражением, вычисляет его и возвращает результат.

**Пример запроса с успешным вычислением выражения:**

```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```

**Пример ответа:**

```json
{
  "result": "6"
}
```

**Пример запроса с ошибкой:**

```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+word"
}'
```

**Пример ответа:**

```json
{
  "error": "Expression is not valid"
}
```

---

## Тестирование

```bash
go test ./...
```