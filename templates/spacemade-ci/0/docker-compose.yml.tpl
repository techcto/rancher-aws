version: '2'

services:
  drone-server:
    image: drone/drone:0.8-alpine
    environment:
      DRONE_BITBUCKET: 'true'
      DRONE_BITBUCKET_CLIENT: ${DRONE_BITBUCKET_CLIENT}
      DRONE_BITBUCKET_SECRET: ${DRONE_BITBUCKET_SECRET}
      DRONE_OPEN: 'true'
      DRONE_HOST: drone-server
      DRONE_SECRET: ${APP_SECRET}
      DRONE_ADMIN: ${DRONE_ADMIN}
      DRONE_DATABASE_DRIVER: mysql
      DRONE_DATABASE_DATASOURCE: ${MYSQL_USER}:${MYSQL_PASSWORD}@mysql/drone?parseTime=true
    volumes:
    - /drone:/var/lib/drone/
    ports:
    - 80:8000/tcp
    labels:
      io.rancher.scheduler.affinity:host_label: drone=server
  drone-agent:
    image: drone/drone:0.8-alpine
    command: agent
    restart: always
    depends_on: [ drone-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=ws://drone-server:8000/ws/broker
      - DRONE_SECRET=${DRONE_SECRET}
  mysql:
    image: mariadb
    command: --sql_mode=""
    environment:
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_USER: '${MYSQL_USER}'
    restart: always