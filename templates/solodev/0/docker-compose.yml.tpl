version: '2'

volumes:

  solodev-client:
    driver_opts:
      repl: '3'
      size: '5'
      shared: true
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
    

services: 

  rancher-lb:
    ports:
     - 80:80
    restart: always
    tty: true
    image: rancher/load-balancer-service
    links:
     - php-fpm:php-fpm
     - apache2:apache2
    stdin_open: true

  php-fpm:
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
    labels:
      io.rancher.container.pull_image: always
    volumes:
      - solodev-client:/var/www/Solodev/clients/solodev
    links:
      - mysql
      - mongo
    depends_on:
      - mysql
    restart: always

  apache2: 
    image: solodev/wcms-apache
    volumes:
      - solodev-client:/var/www/Solodev/clients/solodev
    ports:
      - 80/tcp
      - 443/tcp
    links:
      - php-fpm
    restart: always

  mysql:
    image: mariadb
    command: --sql_mode=""
    environment:
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_USER: '${MYSQL_USER}'
    ports:
      - 3306/tcp
    restart: always
    volumes:
      - solodev-mysql:/var/lib/mysql:rw

  mongo:
    image: 'mongo:3.0'
    environment:
      MONGO_INITDB_ROOT_USERNAME: '${MYSQL_USER}'
      MONGO_INITDB_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
    ports:
      - 27017/tcp
    volumes:
      - solodev-mongo:/data