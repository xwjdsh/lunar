FROM golang:1.17.6-alpine3.15 as builder
WORKDIR /lunar
COPY . .
RUN go build ./cmd/lunar

FROM alpine:3.15.0
LABEL maintainer="iwendellsun@gmail.com"
WORKDIR /workspace
COPY --from=builder /lunar/lunar .
ENTRYPOINT ["/workspace/lunar"]
