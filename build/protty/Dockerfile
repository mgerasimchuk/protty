FROM golang:1.18.9-alpine3.17 as builder

WORKDIR /app
COPY . .

RUN GOSUMDB=off CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -o bin/protty cmd/protty/main.go

FROM alpine:3.17.0

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/protty /bin/protty

EXPOSE 80

CMD ["protty", "start"]
