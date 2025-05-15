# Build stage
FROM golang:1.23 AS build

WORKDIR /app

COPY ./shared ./shared
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poggadaj-tcp

# Runtime stage
FROM alpine:3.20

WORKDIR /app

COPY --from=build /app/poggadaj-tcp .

CMD ["./poggadaj-tcp"]
