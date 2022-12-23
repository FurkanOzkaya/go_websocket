FROM  golang:alpine3.15
RUN apk add git
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . .


EXPOSE 8088

RUN go build -o main . 


CMD ["/app/main"]