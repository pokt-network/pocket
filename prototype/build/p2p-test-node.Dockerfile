FROM golang:1.17.6-bullseye

ENV ADDRESS=
ENV PORT=

# Install dependencies
RUN apk -v --update --no-cache add \
	curl \
	git \
	groff \
	less \
	mailcap \
	gcc \
	libc-dev \
	bash && \
	rm /var/cache/apk/* || true


WORKDIR /go/src/poktp2p

COPY . .

RUN go mod download
RUN go build -o /bin/poktp2p ./cmd/main.go

RUN ls /bin

RUN apt update
RUN apt install telnet

EXPOSE ${PORT}

CMD /bin/poktp2p -address=${ADDRESS}
