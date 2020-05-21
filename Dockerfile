FROM golang:1.14-alpine as builder

RUN mkdir -p /api
COPY ./go.* /api/
RUN cd /api && go mod download

COPY ./*.go /api/
RUN cd /api && go build . && cd / && \
    mv /api/current-challenge /usr/local/bin/ && \
    rm -rf /api

FROM alpine:latest

COPY --from=builder /usr/local/bin/current-challenge /usr/local/bin/

EXPOSE 8080
CMD ["current-challenge"]
