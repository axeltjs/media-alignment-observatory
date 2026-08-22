FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o maoi .

FROM alpine:3.20
RUN apk add --no-cache sqlite-libs ca-certificates
WORKDIR /app
COPY --from=builder /app/maoi .
EXPOSE 8080
CMD ["./maoi"]
