FROM nginx:1.23.3

LABEL Name="icon-repo-client"
LABEL Version="1.0.0"

RUN mkdir -p /opt/iconrepo
COPY ./bundle.js /opt/iconrepo/
COPY ./index.html /opt/iconrepo/
COPY ./nginx.conf /etc/nginx/

EXPOSE 8080
