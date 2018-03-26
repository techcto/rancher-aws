version: '2'

volumes:

  session:

  apistudio-filesystem:
    driver: pxd
    driver_opts:
      repl: '2'
      size: '50'
      shared: true
      io_profile: "cms"

  apistudio-mysql:
    driver: pxd
    driver_opts:
      repl: '3'
      size: '10'
      io_profile: "db"

  apistudio-mongo:
    driver: pxd
    driver_opts:
      repl: '1'
      size: '10'

services:

  php-fpm-lb:
    restart: always
    tty: true
    image: rancher/load-balancer-service
    expose:
      - 9000/tcp
    labels:
      io.rancher.container.agent.role: environmentAdmin
      io.rancher.container.create_agent: 'true'
    links:
      - php-fpm
    stdin_open: true

  php-fpm:
    image: ocoa/apistudio:${APP_BRANCH}
    tty: true
    environment:
      DB_HOST: mysql
      DB_USER: '${MYSQL_USER}'
      DB_PASSWORD: '${MYSQL_PASSWORD}'
      DB_NAME: '${MYSQL_DATABASE}'
      MONGO_HOST: mongo
      OCOA_USER: '${OCOA_USER}'
      OCOA_PASSWORD: '${OCOA_PASSWORD}'
      APP_ENV: '${APP_ENV}'
      APP_DEBUG: 0
      APP_SECRET: '${APP_SECRET}'
      DATABASE_URL: 'mysql://${MYSQL_USER}:${MYSQL_PASSWORD}@mysql:${MYSQL_PORT}/ocoa'
    labels:
      io.rancher.container.network: true
      io.rancher.container.pull_image: always
    volumes:
      - apistudio-filesystem:/var/www/ocoa/fs
      - session:/var/lib/php/session
    links:
      - mysql
      - mongo
    depends_on:
      - mysql
    restart: always

  nginx-lb:
    restart: always
    tty: true
    image: rancher/load-balancer-service
    ports:
      - ${OCOA_PORT}:80
    labels:
      io.rancher.container.agent.role: environmentAdmin
      io.rancher.container.create_agent: 'true'
    links:
      - nginx
    stdin_open: true

  nginx: 
    image: ocoa/apistudio-nginx:${APP_BRANCH}
    labels:
      io.rancher.container.network: true
      io.rancher.container.pull_image: always
    volumes:
      - apistudio-filesystem:/var/www/ocoa/fs
    links:
      - php-fpm-lb:php-fpm
      - react-lb:react
    entrypoint: /usr/local/bin/wait-for-it.sh php-fpm:9000 -t 60 --
    command: ["nginx", "-g", "daemon off;"]
    restart: always

  react-lb:
    restart: always
    tty: true
    image: rancher/load-balancer-service
    labels:
      io.rancher.container.agent.role: environmentAdmin
      io.rancher.container.create_agent: 'true'
    expose:
      - 3000/tcp
    links:
      - react
    stdin_open: true

  react:
    restart: always
    image: ocoa/apistudio-react:${APP_BRANCH}
    labels:
      io.rancher.container.network: true
      io.rancher.container.pull_image: always
    build:
      args: 
        NODE_ENV: ${APP_ENV}

  mysql:
    image: mysql:5.7
    labels:
      io.rancher.container.network: true
    environment:
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_USER: '${MYSQL_USER}'
    tty: true
    stdin_open: true
    volumes:
      - apistudio-mysql:/var/lib/mysql:rw

  mongo:
    image: mongo:3.6
    environment:
      MONGO_INITDB_ROOT_USERNAME: '${MYSQL_USER}'
      MONGO_INITDB_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
    volumes:
      - apistudio-mongo:/data