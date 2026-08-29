# Build stage
FROM golang:1.26-alpine3.24 AS build

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY go.sum go.sum
COPY cmd/poggadaj-http/ cmd/poggadaj-http/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-extldflags "-static"' -o poggadaj-http ./cmd/poggadaj-http

# Runtime stage
FROM alpine:3.24

WORKDIR /app

COPY --from=build /app/poggadaj-http .

CMD ["./poggadaj-http"]
