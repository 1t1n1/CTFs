FROM ghcr.io/stagraphtileful/images:latest@sha256:7e33b501b73da43f71b43db353decdf36d4b1404a4565d31d71fb49d92fba7b2

WORKDIR /app

ENV ENV=local

COPY ./src ./src
COPY ./bin ./bin
COPY ./tests ./tests

RUN /app/bin/build

ENTRYPOINT ["/app/bin/run"]
