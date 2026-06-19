ARG GOLANG_TAG
ARG ALPINE_TAG
FROM --platform=${BUILDPLATFORM} golang:${GOLANG_TAG} AS builder
RUN touch /foobar
FROM alpine:${ALPINE_TAG} AS final
ARG HELLO
RUN echo "..${HELLO}.."
COPY --from=builder /foobar /
ENV HELLO ${HELLO}
ENTRYPOINT [ "/bin/sh", "-c", "echo **${HELLO}**" ]