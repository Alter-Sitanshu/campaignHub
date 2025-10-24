# Build stage
# as i just need the binary after building
FROM golang:1.25-alpine3.22 AS builder

# we specify the working directory for our app
WORKDIR /app

# we copy the files from the current directoy
# to the target i.e /app
COPY . .

# make the executable
RUN go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.0/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY database/migrations ./migrations
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080

CMD [ "/app/main" ]

ENTRYPOINT ["./entrypoint.sh"]