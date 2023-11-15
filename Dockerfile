FROM docker.mofid.dev/golang:1.21.4-alpine3.18 AS build
WORKDIR /artifact
COPY main.go .
COPY go.mod .
COPY go.sum .
RUN go build -o alerting-webhook main.go

FROM docker.mofid.dev/golang:1.21.4-alpine3.18 AS run
WORKDIR /app
COPY --from=build /artifact/alerting-webhook /app/
RUN chmod +x alerting-webhook
EXPOSE 7777
ENTRYPOINT [ "./alerting-webhook" ]
