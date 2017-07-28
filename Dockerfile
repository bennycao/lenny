FROM alpine:3.4

RUN apk add --no-cache ca-certificates

ADD motorhead motorhead
RUN chmod +x motorhead

CMD ["./motorhead"]
