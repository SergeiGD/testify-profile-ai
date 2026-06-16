FROM golang:1.25.7-alpine

WORKDIR /usr/local/src

RUN apk --no-cache add bash gcc musl-dev 

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

COPY . .
COPY ./build/migrations/entrypoint.sh .
COPY ./build/migrations/wait-for-it.sh .

RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]