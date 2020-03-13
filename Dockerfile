FROM        alpine:3.11

LABEL       maintainer="Mittwald CM Service <https://github.com/mittwald>"

ENV         BRUDI_USER="brudi" \
            BRUDI_GID="1000" \
            BRUDI_UID="1000"

COPY        brudi /usr/local/bin/brudi
COPY        --from=restic/restic:0.9.6 /usr/bin/restic /usr/local/bin/restic

RUN         apk add --no-cache --upgrade \
                mongodb-tools \
                mysql-client \
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

USER        ${BRUDI_USER}

ENTRYPOINT  ["brudi"]