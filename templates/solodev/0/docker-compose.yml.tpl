version: '3'

services: 

  solodev:
    image: solodev/wcms
    tty: true
    environment:
      DB_HOST: mysql
      DB_USER: '${MYSQL_USER}'
      DB_PASSWORD: '${MYSQL_PASSWORD}'
      DB_NAME: '${MYSQL_DATABASE}'
      MONGO_HOST: mongo
      SOLODEV_USER: '${SOLODEV_USER}'
      SOLODEV_PASSWORD: '${SOLODEV_PASSWORD}'
    volumes:
      - solodev-client:/var/www/Solodev/clients/solodev
      - solodev:/var/www/Solodev
    links:
      - mysql
      - mongo
    depends_on:
      - mysql
    restart: always

  apache2: 
    image: techcto/docker-solodev-apache2
    volumes:
      - solodev-client:/var/www/Solodev/clients/solodev
      - solodev:/var/www/Solodev
    ports:
      - '${HTTP_PORT}:80'
      - '${HTTPS_PORT}:443'
    links:
      - php-fpm
    depends_on:
      - solodev

  php-fpm:
    image: techcto/docker-php-fpm-7.1
    volumes:
      - solodev-client:/var/www/Solodev/clients/solodev
      - .:/var/www/Solodev
    links:
      - mysql
      - mongo
    restart: always
    depends_on:
      - solodev

  mysql:
    image: mysql:5.7.20
    command: --sql_mode=""
    environment:
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_USER: '${MYSQL_USER}'
    ports:
      - '${MYSQL_PORT}:3306'
    restart: always
    volumes:
      - solodev-mysql:/var/lib/mysql:rw

  mongo:
    image: 'mongo:3.0'
    environment:
      MONGO_INITDB_ROOT_USERNAME: '${MYSQL_USER}'
      MONGO_INITDB_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
    ports:
      - '${MONGODB_PORT}:27017'
    volumes:
      - solodev-mongo:/data

volumes:

  solodev:

  solodev-client:
    driver_opts:
      repl: '3'
      size: '5'
    driver: pxd

  solodev-mysql:
    driver_opts:
      repl: '3'
      size: '5'
    driver: pxd

  solodev-mongo:
    driver_opts:
      repl: '3'
      size: '5'
    driver: pxd
