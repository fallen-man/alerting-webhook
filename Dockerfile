FROM docker.mofid.dev/golang:1.21.4-alpine3.18 AS build
WORKDIR /artifact
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN  go build -o alerting-webhook main.go

FROM scratch
WORKDIR /go/bin
COPY --from=build /artifact/alerting-webhook /go/bin/alerting-webhook
EXPOSE 7777
ENTRYPOINT [ "./alerting-webhook" ]