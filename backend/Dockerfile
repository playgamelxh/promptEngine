# Build stage
FROM docker.m.daocloud.io/library/golang:1.23-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod ./
# COPY go.sum ./ # go.sum doesn't exist yet as we haven't run go mod tidy
RUN go mod download || true # Ignore error if go.sum missing

COPY . .

RUN go mod tidy
RUN go build -o main main.go

# Run stage
FROM docker.m.daocloud.io/library/alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
