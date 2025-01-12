FROM ubuntu:latest

RUN apt-get -y install coreutils grep findutils systemd bzip2 make rpm-build build-essential
RUN curl -L -o go1.21.1.linux-amd64.tar.gz https://go.dev/dl/go1.21.1.linux-amd64.tar.gz && rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz
