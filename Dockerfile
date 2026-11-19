#syntax=docker/dockerfile:1
FROM alpine:latest
ARG TARGETARCH
RUN apk --no-cache add ca-certificates
ENV BASE_DIR=/ga2tg
WORKDIR ${BASE_DIR}
COPY ./.env /ga2tg/
COPY ./settings/ /ga2tg/settings/
COPY ./bin/ga2tg-linux-${TARGETARCH} /ga2tg/main
EXPOSE 8080
ENTRYPOINT ["/ga2tg/main"]