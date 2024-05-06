FROM golang:latest
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
COPY .env /app

ENV GIN_MODE=release
ENV GOOS=linux
RUN go build -ldflags="-s -w" -buildvcs=false -o /khoomi ./cmd/khoomi

EXPOSE 8080
ENTRYPOINT ["/khoomi"]
