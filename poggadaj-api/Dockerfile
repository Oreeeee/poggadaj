# Build stage
FROM golang:1.24.6-alpine AS build

WORKDIR /app

COPY ./shared ./shared
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poggadaj-api

# Runtime stage
FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/poggadaj-api .

CMD ["./poggadaj-api"]
