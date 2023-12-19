FROM ubuntu:22.04

LABEL Name="iconrepo"
LABEL Version="1.0.0"

RUN useradd -m --shell /bin/false iconrepo

RUN apt-get update && apt-get install -y git curl && apt-get clean
RUN mkdir -p /opt/iconrepo
WORKDIR /opt/iconrepo
COPY iconrepo/ ./
RUN chmod +x iconrepo
RUN ls -la /opt/
RUN mkdir "/data"
EXPOSE 8080

USER iconrepo

CMD [ "/opt/iconrepo/iconrepo" ]
