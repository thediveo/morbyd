FROM --platform=${BUILDPLATFORM} golang:1-alpine AS builder
FROM alpine AS final
ARG HELLO
RUN echo "..${HELLO}.."
ENV HELLO ${HELLO}
ENTRYPOINT [ "/bin/sh", "-c", "echo **${HELLO}**" ]