FROM golang:1.15-alpine

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY . .

CMD [ "go", "run", "main.go" ]
