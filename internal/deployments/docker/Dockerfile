FROM ubuntu:20.04
RUN apt-get update && apt-get install -y git
RUN mkdir -p /opt/igo-repo
WORKDIR /opt/igo-repo
COPY igo-repo/ ./
RUN chmod +x igo-repo
RUN ls -la /opt/
RUN mkdir "/data"
EXPOSE 8080
CMD [ "/opt/igo-repo/igo-repo" ]
