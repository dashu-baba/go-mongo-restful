FROM golang:alpine

WORKDIR /opt/app

COPY go.* /opt/app/

RUN go mod download

COPY . .

RUN go build -o anabel.exe main.go

EXPOSE 4001
EXPOSE 4201