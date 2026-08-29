# Build stage
FROM golang:1.26-alpine3.24 AS build

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY go.sum go.sum
COPY cmd/poggadaj-frontend/ cmd/poggadaj-frontend/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-extldflags "-static"' -o poggadaj-frontend ./cmd/poggadaj-frontend

# Runtime stage
FROM alpine:3.24

WORKDIR /app

COPY ./cmd/poggadaj-frontend/static ./static
COPY ./cmd/poggadaj-frontend/i18n ./i18n
COPY ./cmd/poggadaj-frontend/views ./views
COPY --from=build /app/poggadaj-frontend .

CMD ["./poggadaj-frontend"]
