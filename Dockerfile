FROM golang:1.25-alpine AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/aifriends ./cmd/server

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/aifriends /app/aifriends
COPY backend/configs /app/configs
COPY backend/static /app/static
COPY backend/documents /app/documents
COPY backend/media /app/default-media

ENV CONFIG_PATH=/app/configs/config.yaml

EXPOSE 8080

CMD ["sh", "-c", "mkdir -p /app/media/user/photos && if [ ! -f /app/media/user/photos/default.jpg ]; then cp /app/default-media/user/photos/default.jpg /app/media/user/photos/default.jpg; fi && exec /app/aifriends"]
