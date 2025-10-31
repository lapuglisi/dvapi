FROM archlinux:latest

WORKDIR /app

COPY go.mod main.go app.go go.sum dvapi.db.dist ./
COPY database/ ./database/
COPY model/ ./model/
COPY http/ ./http/

RUN mv ./dvapi.db.dist ./dvapi.db

RUN pacman-key --init && pacman-key --populate archlinux

RUN pacman --noconfirm -Syyu && pacman --noconfirm -S base-devel go

RUN ls -la ./

RUN go build ./

EXPOSE 9098

CMD ["./dvapi"]
