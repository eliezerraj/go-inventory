# docker build -t go-inventory .
# docker run -dit --name go-inventory -p 7000:7000 go-inventory

FROM golang:1.24 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .
RUN go mod tidy

WORKDIR /app/cmd
RUN go build -o go-inventory -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-inventory .

WORKDIR /var/pod/secret

#RUN echo -n "postgres" > /var/pod/secret/username # for testing only
#RUN echo -n "postgres" > /var/pod/secret/password # for testing only
#COPY --from=builder /app/cmd/.env . #for testing only

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-inventory"]