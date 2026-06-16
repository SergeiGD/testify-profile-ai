FROM golang:1.25.7-alpine

WORKDIR /usr/local/src

RUN apk --no-cache add gcc musl-dev

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]
