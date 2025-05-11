FROM golang:latest AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler .

FROM alpine:latest

RUN apk update && apk add --no-cache tzdata


WORKDIR /app

COPY --from=builder /app/scheduler .

COPY web ./web

EXPOSE 7540

ENTRYPOINT ["/app/scheduler"]
