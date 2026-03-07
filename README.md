# Mommy AI API

Simple Go API using [Gin](https://gin-gonic.com/) with an OpenAI-compatible LLM (e.g. Yandex Cloud AI) for chat.

## Run

```bash
go mod tidy
# On macOS, use static build to avoid "missing LC_UUID" dyld error (Go 1.21 + newer macOS):
CGO_ENABLED=0 go run .
# Or build then run:
CGO_ENABLED=0 go build -o mommy-ai-api . && ./mommy-ai-api
```

Server listens on `http://localhost:8080`.

## Endpoints

- `GET /health` — health check
- `GET /api/hello` — sample greeting
- `POST /api/chat` — chat with the LLM

### Chat

Request (single message):

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}'
```

Request (full message history, OpenAI-style):

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"messages": [{"role": "user", "content": "Hi"}, {"role": "assistant", "content": "Hello!"}, {"role": "user", "content": "What is 2+2?"}]}'
```

Response:

```json
{"reply": "The model's reply text..."}
```
