FROM debian:bookworm-slim AS build-pannellum
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates curl unzip && rm -rf /var/lib/apt/lists/*
WORKDIR /mnt
COPY ./build-pannellum.sh ./build.sh
RUN ./build.sh

FROM debian:bookworm-slim AS result
COPY --from=build-pannellum /result/ /result/
RUN tar -vcf /result.tar -C /result ./
