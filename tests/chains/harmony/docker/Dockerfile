# Build atlas in a stock Go builder container
FROM golang:1.18 as builder

#ARG MAP_TAG

RUN apt install make gcc git

RUN cd /root && git clone https://github.com/neoiss/atlas.git -b update-max-code-size --depth 1

WORKDIR /root/atlas

RUN make atlas

# Pull atlas into a second stage deploy ubuntu container
FROM ubuntu:latest

RUN apt update && apt install -y curl

COPY --from=builder /root/atlas/build/bin/atlas /usr/local/bin/
RUN chmod +x /usr/local/bin/atlas

EXPOSE 7445
ENTRYPOINT ["atlas", "--datadir", "/root/atlas/datadir", "--http", "--http.corsdomain", "*", "--http.vhosts", "*", "--single"]
