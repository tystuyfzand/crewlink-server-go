FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY . .

RUN go get -u github.com/gobuffalo/packr/v2/... && packr2

RUN go build -o crewlink-server cmd/main.go

FROM alpine

COPY --from=builder /build/crewlink-server /crewlink-server

EXPOSE 9736

CMD ["/crewlink-server"]