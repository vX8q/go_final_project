FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/scheduler .
COPY web ./web

RUN adduser -D scheduleruser
USER scheduleruser

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
ENV TODO_PASSWORD=""

VOLUME /data
EXPOSE $TODO_PORT

CMD ["./scheduler"]