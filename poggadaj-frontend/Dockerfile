# Build stage
FROM golang:1.26-alpine3.23 AS build

WORKDIR /app

COPY ./shared ./shared
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poggadaj-frontend

# Runtime stage
FROM alpine:3.23

WORKDIR /app

COPY ./html ./html
COPY --from=build /app/poggadaj-frontend .

CMD ["./poggadaj-frontend"]
