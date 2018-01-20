FROM node:9.3.0-alpine

# Override the base log level (info).
ENV NPM_CONFIG_LOGLEVEL warn

# Install and configure `serve`.
RUN npm install -g serve

# Install all dependencies of the current project.
COPY client/package.json ./package.json
RUN yarn install

# Copy all local files into the image.
COPY client/. .

# Build for dev.
#RUN npm run start --dev

# Build for production.
RUN npm run build --production

CMD serve -s build
EXPOSE 5000
