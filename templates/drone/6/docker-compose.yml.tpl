version: '2'
services:
  mysql:
    image: mariadb
    command: --sql_mode=""
    environment:
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_USER: '${MYSQL_USER}'
    restart: always
    volumes:
      - wxp-mysql:/var/lib/mysql:rw
  drone-agent:
    image: drone/agent:0.8-alpine
    environment:
      DRONE_SERVER: DRONE_SERVER=drone-server:8000
      DRONE_SECRET: ${drone_secret}
{{- if (.Values.http_proxy)}}
      HTTP_PROXY: ${http_proxy}
      http_proxy: ${http_proxy}
{{- end}}
{{- if (.Values.https_proxy)}}
      HTTPS_PROXY: ${https_proxy}
      https_proxy: ${https_proxy}
{{- end}}
{{- if (.Values.no_proxy)}}
      NO_PROXY: ${no_proxy}
      no_proxy: ${no_proxy}
{{- end}}
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    links:
      - drone-lb:drone-server
    depends_on: [ drone-server ]
    command:
      - agent
    labels:
      io.rancher.scheduler.affinity:host_label_ne: drone:server
      io.rancher.scheduler.global: 'true'
  drone-server:
    image: drone/drone:0.8-alpine
    links:
      - mysql
    ports:
      - 80:8000/tcp
    environment:
      DRONE_HOST: http://drone-server
      GIN_MODE: ${gin_mode}
{{- if (.Values.drone_debug)}}
      DRONE_DEBUG: '${drone_debug}'
{{- end}}
      DRONE_SECRET: ${drone_secret}
      DRONE_OPEN: ${drone_open}
{{- if (.Values.drone_admin)}}
      DRONE_ADMIN: ${drone_admin}
{{- end}}
{{- if (.Values.drone_orgs)}}
      DRONE_ORGS: ${drone_orgs}
{{- end}}
{{- if eq .Values.drone_driver "GitHub"}}
      DRONE_GITHUB: true
      DRONE_GITHUB_CLIENT: ${drone_driver_client}
      DRONE_GITHUB_SECRET: ${drone_driver_secret}
{{- end}}
{{- if eq .Values.drone_driver "Bitbucket Cloud"}}
      DRONE_BITBUCKET: true
      DRONE_BITBUCKET_CLIENT: ${drone_driver_client}
      DRONE_BITBUCKET_SECRET: ${drone_driver_secret}
{{- end}}
{{- if eq .Values.drone_driver "Bitbucket Server"}}
      DRONE_STASH: true
      DRONE_STASH_GIT_USERNAME: ${drone_driver_user}
      DRONE_STASH_GIT_PASSWORD: ${drone_driver_password}
      DRONE_STASH_CONSUMER_KEY: ${drone_driver_client}
      DRONE_STASH_CONSUMER_RSA_STRING: ${drone_driver_secret}
      DRONE_STASH_URL: ${drone_driver_url}
{{- end}}
{{- if eq .Values.drone_driver "GitLab"}}
      DRONE_GITLAB: true
      DRONE_GITLAB_CLIENT: ${drone_driver_secret}
      DRONE_GITLAB_SECRET: ${drone_driver_secret}
      DRONE_GITLAB_URL: ${drone_driver_url}
{{- end}}
{{- if eq .Values.drone_driver "Gogs"}}
      DRONE_GOGS: true
      DRONE_GOGS_URL: ${drone_driver_url}
{{- end}}
      DRONE_DATABASE_DRIVER: mysql
      DRONE_DATABASE_DATASOURCE: ${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(mysql:3306)/drone?parseTime=true
{{- if (.Values.http_proxy)}}
      HTTP_PROXY: ${http_proxy}
      http_proxy: ${http_proxy}
{{- end}}
{{- if (.Values.https_proxy)}}
      HTTPS_PROXY: ${https_proxy}
      https_proxy: ${https_proxy}
{{- end}}
{{- if (.Values.no_proxy)}}
      NO_PROXY: ${no_proxy}
      no_proxy: ${no_proxy}
{{- end}}
    labels:
      io.rancher.scheduler.affinity:host_label: drone:server
  drone-lb:
    restart: always
    tty: true
    image: rancher/load-balancer-service
    ports:
      - ${host_port}:${host_port}
    labels:
      io.rancher.container.agent.role: environmentAdmin
      io.rancher.container.create_agent: 'true'
    links:
      - drone-server
    stdin_open: true
