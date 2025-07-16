FROM node:18-bullseye AS build-openseadragon
WORKDIR /mnt
COPY ./build-openseadragon.sh ./build.sh
RUN ./build.sh

FROM node:18-bullseye AS build-openseadragon-flat-toolbar-icons
WORKDIR /mnt
COPY ./build-openseadragon-flat-toolbar-icons.sh ./build.sh
RUN ./build.sh

FROM debian:bookworm-slim AS result
COPY --from=build-openseadragon /result/ /result/
COPY --from=build-openseadragon-flat-toolbar-icons /result/ /result/
RUN tar -vcf /result.tar -C /result ./
