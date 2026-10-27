FROM docker.io/library/alpine:3.22
RUN apk add --no-cache openssh-server rsync \
  && echo 'rsyncbackup:x:0:0:backup:/root:/bin/sh' >> /etc/passwd \
  && echo 'rsyncbackup:*::0:::::' >> /etc/shadow \
  && rm -f /etc/ssh/ssh_host_*
COPY sshd_config /etc/ssh/sshd_config
COPY rsync-sender.sh /usr/local/bin/rsync-sender
RUN chmod 0755 /usr/local/bin/rsync-sender
EXPOSE 22
ENTRYPOINT ["/usr/sbin/sshd", "-D", "-e"]
