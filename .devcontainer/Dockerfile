FROM mcr.microsoft.com/devcontainers/base:bookworm
RUN sudo apt update

# habits
RUN sudo apt install -y vim

# vips
RUN sudo apt install -y libvips libvips-tools

# exiftool
RUN mkdir /tmp/exiftool \
  && cd /tmp/exiftool \
  && wget https://exiftool.org/Image-ExifTool-13.31.tar.gz \
  && tar -xzf Image-ExifTool-*.tar.gz \
  && cd Image-ExifTool-* \
  && perl Makefile.PL \
  && sudo make install \
  && rm -fr /tmp/exiftool

# tailwindcss
RUN sudo wget -O /usr/local/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.10/tailwindcss-linux-arm64 \
  && sudo chmod +x /usr/local/bin/tailwindcss

# build scripts
RUN sudo apt install -y xxd brotli zstd
