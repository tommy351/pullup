FROM gcr.io/distroless/base
ARG BINARY_NAME
COPY $BINARY_NAME /pullup
CMD ["/pullup"]
