FROM debian:trixie-slim

WORKDIR /app

COPY go.mod main.go app.go go.sum dvapi.db.dist ./
COPY database/ ./database/
COPY model/ ./model/
COPY http/ ./http/

RUN mv ./dvapi.db.dist ./dvapi.db

RUN apt-get -y update && apt-get -y install build-essential ca-certificates

RUN ls -la ./

RUN go build ./

EXPOSE 9098

ENV DVAPI_PORT 9098
ENV DVAPI_HOST "0.0.0.0"

CMD ["/bin/sh", "-c", "./dvapi -port ${DVAPI_PORT:-9098} -host ${DVAPI_HOST:-0.0.0.0}"]
