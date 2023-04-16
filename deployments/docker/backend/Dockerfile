FROM ubuntu:22.04

LABEL Name="icon-repo"
LABEL Version="1.0.0"

RUN useradd -m --shell /bin/false iconrepo

RUN apt-get update && apt-get install -y git && apt-get clean
RUN mkdir -p /opt/igo-repo
WORKDIR /opt/igo-repo
COPY igo-repo/ ./
RUN chmod +x igo-repo
RUN ls -la /opt/
RUN mkdir "/data"
EXPOSE 8080

USER iconrepo

CMD [ "/opt/igo-repo/igo-repo" ]
