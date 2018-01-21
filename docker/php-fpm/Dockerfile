FROM php:7.2.0-fpm-alpine3.7

RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

RUN set -xe

# Default PHP Configurations values.
ENV PHP_MEMORY_LIMIT 64M
ENV PHP_TIMEZONE=UTC
ENV PHP_REALPATH_CACHE_SIZE 2M

# FPM Configurations.
ENV PHP_FPM_MAX_CHILDREN 5
ENV PHP_FPM_START_SERVERS 2
ENV PHP_FPM_MIN_SPARE_SERVERS 1
ENV PHP_FPM_MAX_SPARE_SERVERS 3

ENV XDEBUG_VERSION 2.6.0alpha1
ENV MEMCACHE_VERSION 3.0.4
ENV PHPREDIS_VERSION 3.1.4
ENV PHP_EXTRA_DEPS libxml2-dev icu-dev libmemcached-dev cyrus-sasl-dev libpng libjpeg-turbo freetype-dev libpng-dev libjpeg-turbo-dev
RUN apk update && \
    apk add ${PHP_EXTRA_DEPS} ${PHPIZE_DEPS} && \
    docker-php-ext-configure gd --with-gd --with-freetype-dir=/usr/include/ --with-png-dir=/usr/include/ --with-jpeg-dir=/usr/include/ && \
    apk add ${PHP_EXTRA_DEPS} ${PHPIZE_DEPS} gettext && \
    docker-php-ext-install soap gd mbstring pdo pdo_mysql zip intl pcntl && \
    pecl install xdebug-${XDEBUG_VERSION} && \
    pecl install redis-${PHPREDIS_VERSION} && \
    pecl install igbinary memcached-${MEMCACHE_VERSION} --disable-memcached-sasl && \
    apk del ${PHPIZE_DEPS} && \
    mkdir -p /tmp/blackfire && \
    curl -A "Docker" -L https://blackfire.io/api/v1/releases/client/linux_static/amd64 | tar zxp -C /tmp/blackfire && \
    mv /tmp/blackfire/blackfire /usr/bin/blackfire && \
    rm -Rf /tmp/blackfire && \
    rm -rf /var/cache/apk/*

# Enable XDebug
RUN docker-php-ext-enable xdebug
ENV XDEBUGINI_PATH=/usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini
RUN echo "zend_extension="`find /usr/local/lib/php/extensions/ -iname 'xdebug.so'` > $XDEBUGINI_PATH
COPY docker/php/xdebug.ini /tmp/xdebug.ini
RUN cat /tmp/xdebug.ini >> $XDEBUGINI_PATH

# Install Mailhog.
ENV MAILHOG_VERSION v0.1.9
RUN curl -fSL "https://github.com/mailhog/MailHog/releases/download/${MAILHOG_VERSION}/MailHog_linux_amd64" -o /usr/local/bin/mailhog \
    && chmod +x /usr/local/bin/mailhog

#Install Common Libs
RUN apk add --no-cache zip unzip wget curl openssl git

#Install Composer
RUN curl -sS https://getcomposer.org/installer | \
    php -- --install-dir=/usr/bin/ --filename=composer

# https://getcomposer.org/doc/03-cli.md#composer-allow-superuser
ENV COMPOSER_ALLOW_SUPERUSER 1

RUN composer global require "hirak/prestissimo:^0.3" --prefer-dist --no-progress --no-suggest --optimize-autoloader --classmap-authoritative \
	&& composer clear-cache

#Remove useless config
RUN rm -rf /etc/php7/php-fpm.d

WORKDIR /var/www

#Set port
EXPOSE 9000

CMD ["php-fpm"]