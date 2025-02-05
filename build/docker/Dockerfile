FROM        alpine:3.21

LABEL       maintainer="Mittwald CM Service <https://github.com/mittwald>"

ENV         BRUDI_USER="brudi" \
            BRUDI_GID="1000" \
            BRUDI_UID="1000"

COPY        brudi /usr/local/bin/brudi

COPY        --from=restic/restic:0.11.0 /usr/bin/restic /usr/local/bin/restic
COPY        --from=redis:alpine /usr/local/bin/redis-cli /usr/local/bin/redis-cli

RUN         apk add --no-cache --upgrade \
                mongodb-tools \
                mysql-client \
                postgresql-client \
                gcompat \
            && \
            addgroup \
                -S "${BRUDI_USER}" \
                -g "${BRUDI_GID}" \
            && \
            adduser \
                -u "${BRUDI_UID}" \
                -S \
                -G "${BRUDI_USER}" \
                "${BRUDI_USER}"

RUN         curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg && \
            sudo chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg && \
            echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list && \
            sudo apt-get update && \
            sudo apt-get install redis-tools=6:6.2.15-1rl1~noble1

USER        ${BRUDI_USER}

ENTRYPOINT  ["brudi"]