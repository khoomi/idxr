FROM golang:latest
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
COPY .env /app

ENV GIN_MODE=release
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -buildvcs=false -o /khoomi

EXPOSE 8080
ENTRYPOINT ["/khoomi"]
