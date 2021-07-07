FROM node:carbon
WORKDIR /usr/src/app
COPY dist/ ./
WORKDIR /usr/src/app/backend
RUN npm install --only=production
VOLUME ["/data"]
EXPOSE 8080
CMD [ "node", "app.js" ]
