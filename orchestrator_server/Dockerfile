FROM golang:1.20

WORKDIR /orchestrator_server

COPY go.mod .
COPY . .

CMD ["go", "run", "main.go"]