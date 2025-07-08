FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o retro_aim_server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/retro_aim_server /app/

EXPOSE 8080 5190 5191 5192 5193 5194 5195 5196 5197

CMD ["/app/retro_aim_server"]
