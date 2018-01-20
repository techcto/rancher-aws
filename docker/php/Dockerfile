FROM solodev/php-fpm-alpine

# dev ONLY
RUN apk add --no-cache --virtual .build-deps \
    $PHPIZE_DEPS \
    && pecl install xdebug \
    && docker-php-ext-enable xdebug
ENV XDEBUGINI_PATH=/usr/local/etc/php/conf.d/docker-php-ext-xdebug.ini
RUN echo "zend_extension="`find /usr/local/lib/php/extensions/ -iname 'xdebug.so'` > $XDEBUGINI_PATH
COPY docker/php/xdebug.ini /tmp/xdebug.ini
RUN cat /tmp/xdebug.ini >> $XDEBUGINI_PATH

COPY composer.lock ./
COPY composer.json ./

WORKDIR /var/www

# Note: Production only - composer install with --classmap-authoritative
RUN mkdir -p fs var/cache var/logs var/sessions \
    && composer install --prefer-dist --no-dev --no-progress --no-suggest --no-interaction \
    && composer clear-cache \
    && chown -Rf www-data var \
    && chown -Rf www-data fs

RUN mkdir -p var/jwt
RUN openssl genrsa -aes256 -passout pass:solodev -out var/jwt/private.pem 4096
RUN openssl rsa -pubout -passin pass:solodev -in var/jwt/private.pem -out var/jwt/public.pem

#Copy special configs
COPY docker/php/php.ini /etc/php7/php.ini
COPY docker/php/php-fpm.conf /etc/php7/php-fpm.conf

#Set port
EXPOSE 9000

COPY config ./config
COPY src ./src
COPY bin ./bin

CMD ["php-fpm"]