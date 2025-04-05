FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /auth-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /
COPY --from=builder /auth-service /auth-service

ENV SQLITE_PATH=/data/auth.db
ENV JWT_SECRET=your-256-bit-secret
ENV PORT=8080

VOLUME /data
EXPOSE 8080
CMD ["/auth-service"]