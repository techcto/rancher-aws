FROM php:7.2.0-fpm-alpine3.7

RUN echo "http://dl-4.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories

ENV XDEBUG_VERSION 2.4.1
ENV MEMCACHE_VERSION 2.2.0
ENV TIDY_VERSION=5.1.25

RUN apk add --no-cache \
		bash \
		sed

RUN set -ex; \
	\
	apk add --no-cache --virtual .build-deps \
		autoconf \
		build-base \
		bzip2-dev \
		gettext-dev \
		icu-dev \
		imagemagick-dev \
		imap-dev \
		libjpeg-turbo-dev \
		libmcrypt-dev \
		libmemcached-dev \
		libpng-dev \
		libssh2-dev \
		libtool \
		libxml2-dev \
		libxslt-dev \
		openldap-dev \
		pcre-dev \
        make zip unzip wget curl cmake \
	; 

RUN docker-php-ext-configure gd --with-png-dir=/usr --with-jpeg-dir=/usr; \
	docker-php-ext-install \
		bcmath \	
		bz2 \
		calendar \
		dba \
		exif \
		gd \
		gettext \
		imap \
		intl \
		ldap \
		mcrypt \
		mysqli \
		opcache \
		pdo_mysql \
		shmop \
		soap \
		sockets \
		sysvmsg \
		sysvsem \
		sysvshm \
		wddx \
		xmlrpc \
		xsl \ 
		zip \
	; \
	pecl install \
        apcu \
		igbinary \
		imagick \
		memcached \
		msgpack \
        mongodb \
        oauth \
		redis \
	; \
	echo "\n" | pecl install ssh2-1.0; \
	docker-php-ext-enable --ini-name pecl.ini \
        apcu \
		igbinary \
		imagick \
		memcached \
		msgpack \
        mongodb \
        oauth \
		redis \
		ssh2 \
	; \
	runDeps="$( \
		scanelf --needed --nobanner --recursive \
			/usr/local/lib/php/extensions \
			| awk '{ gsub(/,/, "\nso:", $2); print "so:" $2 }' \
			| sort -u \
			| xargs -r apk info --installed \
			| sort -u \
	)"; \
	apk add --virtual $runDeps; \
	rm -rf /tmp/pear/;

COPY php.ini /usr/local/etc/php/php.ini
COPY php-fpm.conf /usr/local/etc/php-fpm.conf

#Install Composer
RUN curl -sS https://getcomposer.org/installer | \
    php -- --install-dir=/usr/bin/ --filename=composer

# https://getcomposer.org/doc/03-cli.md#composer-allow-superuser
ENV COMPOSER_ALLOW_SUPERUSER 1

RUN composer global require "hirak/prestissimo:^0.3" --prefer-dist --no-progress --no-suggest --optimize-autoloader --classmap-authoritative \
	&& composer clear-cache

#Remove useless config
RUN rm -rf /etc/php7/php-fpm.d