FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./server/main.go

ARG PORT=8080

EXPOSE ${PORT}

CMD sh -c "./main -port=${PORT}"
