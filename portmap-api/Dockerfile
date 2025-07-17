FROM artifacts.developer.gov.bc.ca/docker-remote/golang:1.16-alpine as builder

RUN mkdir -p /build
WORKDIR /build

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .

RUN go build -o portmap-api

FROM artifacts.developer.gov.bc.ca/docker-remote/alpine:latest

WORKDIR /app

COPY --from=builder /build/portmap-api /app/

ENV HOME=/app \
    USER_NAME=golang \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd \
  && chown -R ${USER_UID}:0 ${HOME} \
  && chmod -R ug+rwx ${HOME}

USER golang

CMD ["./portmap-api"]
