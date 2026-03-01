FROM mcr.microsoft.com/devcontainers/base:jammy
RUN sudo apt update

# habits
RUN sudo apt install -y vim

# vips
# RUN sudo ./private/scripts/libvips/install.sh

# exiftool
RUN mkdir /tmp/exiftool \
  && cd /tmp/exiftool \
  && wget -O Image-ExifTool.tar.gz 'https://cytranet-dal.dl.sourceforge.net/project/exiftool/Image-ExifTool-13.52.tar.gz?viasf=1' \
  && tar -xzf Image-ExifTool.tar.gz \
  && cd Image-ExifTool-* \
  && perl Makefile.PL \
  && sudo make install \
  && rm -fr /tmp/exiftool

# tailwindcss
RUN sudo wget -O /usr/local/bin/tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/download/v4.1.10/tailwindcss-linux-arm64 \
  && sudo chmod +x /usr/local/bin/tailwindcss

# build scripts
RUN sudo apt install -y xxd brotli zstd
