FROM golang:1.13 as builder
COPY . /coap-hooks-router
WORKDIR /coap-hooks-router
ENV CGO_ENABLED 0
RUN go build -o server .

FROM scratch as release
ARG ADMIN_BEARER
WORKDIR /
VOLUME /dbdata
COPY --from=builder /coap-hooks-router/server /server
ENTRYPOINT ["/server"]
