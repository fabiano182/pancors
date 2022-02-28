FROM golang:1.17-alpine3.15 as builder

WORKDIR go/src/github.com/fabiano182/pancors

COPY . .

RUN go build -o pancors ./cmd/pancors/main.go


FROM alpine:3.15

WORKDIR /app/

COPY --from=builder /go/src/githu.com/fabiano182/pancors/pancors ./

EXPOSE 8080

CMD [ "./pancors" ]
