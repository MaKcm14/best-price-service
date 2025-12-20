# Creating the PE-app file.
FROM golang:1.23-alpine3.21 AS builder

LABEL maintainer="maksimacx50@gmail.com"

WORKDIR /price-service
COPY . .

RUN go mod tidy

WORKDIR /price-service/cmd/app
RUN go build main.go


# Starting the installing.
FROM chromedp/headless-shell:latest

WORKDIR /cmd/app
COPY --from=builder /price-service/cmd/app/main .

WORKDIR /
COPY --from=builder /price-service/.env .

RUN apt-get update && \ 
	apt-get install -y \
	dumb-init \
	ca-certificates \
	xmlsec1 \
	procps

VOLUME /logs

EXPOSE 8080

WORKDIR /cmd/app

ENTRYPOINT [ "./main" ]
