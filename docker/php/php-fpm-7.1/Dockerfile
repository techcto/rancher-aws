FROM php:7.1-fpm-alpine

RUN set -xe

# FPM Configurations.
ENV PHP_FPM_MAX_CHILDREN 5
ENV PHP_FPM_START_SERVERS 2
ENV PHP_FPM_MIN_SPARE_SERVERS 1
ENV PHP_FPM_MAX_SPARE_SERVERS 3

ENV XDEBUG_VERSION 2.4.1
ENV MEMCACHE_VERSION 2.2.0
ENV TIDY_VERSION=5.1.25

RUN apk add --no-cache --virtual .phpize-deps-1 $PHPIZE_DEPS \
    freetype-dev libpng-dev libjpeg-turbo-dev libxml2-dev imagemagick-dev libtool \
    make pcre-dev zip unzip wget curl openssl make cmake \
    openssl openssl-dev libmcrypt-dev \
    gettext gettext-libs gettext-dev \
    && apk add --no-cache --virtual .imagick-runtime-deps imagemagick \
    && apk add --no-cache --virtual .other-php-ext-deps libmcrypt libxslt-dev freetds freetds-dev bzip2-dev

RUN mkdir -p /usr/local/src \
    && cd /usr/local/src \
    && curl -q https://codeload.github.com/htacg/tidy-html5/tar.gz/$TIDY_VERSION | tar -xz \
    && cd tidy-html5-$TIDY_VERSION/build/cmake \
    && cmake ../.. && make install \
    && ln -s tidybuffio.h ../../../../include/buffio.h \
    && cd /usr/local/src \
    && rm -rf /usr/local/src/tidy-html5-$TIDY_VERSION \
    && pecl install mongodb \
    && echo "extension=mongodb.so" > /usr/local/etc/php/conf.d/mongodb.ini \
    && pecl install mongo \
    && echo "extension=mongo.so" > /usr/local/etc/php/conf.d/mongo.ini \
    && pecl install imagick \
    && docker-php-ext-enable imagick \
    && pecl install oauth \
    && docker-php-ext-enable oauth \
    && pecl clear-cache \
    && docker-php-ext-configure gd \
        --with-gd \
        --with-freetype-dir=/usr/include/ \
        --with-png-dir=/usr/include/ \
        --with-jpeg-dir=/usr/include/ \
    && docker-php-ext-install gd \
    && docker-php-ext-install tidy \
    && docker-php-ext-install bz2 \
    && docker-php-ext-install bcmath \
    && docker-php-ext-install calendar \
    && docker-php-ext-install exif \
    && docker-php-ext-install mcrypt \
    && docker-php-ext-install mysqli \
    && docker-php-ext-install opcache \
    && docker-php-ext-install pdo_dblib \
    && docker-php-ext-install pdo_mysql \
    && docker-php-ext-install shmop \
    && docker-php-ext-install sockets \
    && docker-php-ext-install soap \
    && docker-php-ext-install sysvmsg \
    && docker-php-ext-install wddx \
    && docker-php-ext-install xsl \
    && docker-php-ext-install zip \
    && docker-php-ext-install gettext \
	&& pecl install apcu \
	&& echo "extension=apcu.so" > /usr/local/etc/php/conf.d/apcu.ini \
	&& apk del autoconf g++ libtool make pcre-dev \
    && apk del .phpize-deps-1 \
    && rm -rf /tmp/* /var/cache/apk/*

COPY php.ini /usr/local/etc/php/conf.d/php.ini
COPY php-fpm.conf /usr/local/etc/

WORKDIR /var/www

EXPOSE 9000

CMD ["php-fpm"]