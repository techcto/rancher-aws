FROM 'httpd:alpine'

# Solodev Config
RUN echo "Include conf/extra/solodev.conf" >> /usr/local/apache2/conf/httpd.conf
COPY solodev.conf /usr/local/apache2/conf/extra/solodev.conf
RUN mkdir -p /var/www/Solodev/clients/solodev/Vhosts
RUN mkdir -p /var/www/Solodev/clients/solodev/s.Vhosts