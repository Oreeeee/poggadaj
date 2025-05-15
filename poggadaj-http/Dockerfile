# Build stage
FROM golang:1.23 AS build

WORKDIR /app

COPY ./shared ./shared
COPY go.mod .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poggadaj-http

# Runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/poggadaj-http .

CMD ["./poggadaj-http"]
