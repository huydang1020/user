FROM alpine:3.14

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root

# Copy binary đã build sẵn và assets
COPY user .
COPY assets ./assets

EXPOSE 6000 6001

CMD ["./user", "start"]