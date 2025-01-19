FROM alpine:3

WORKDIR /app/tbsign

ARG TARGETARCH
COPY tbsign_go.musl.${TARGETARCH}/tbsign_go /app/tbsign/

ENV tc_db_path="/app/tbsign/db/tbsign_go.db" \
    tc_api="true" \
    tc_fe="true" \
    tc_address=":1323"

RUN echo -e "#!/bin/sh\nsh -c /app/tbsign/tbsign_go" > /run.sh &&\
    chmod +x /app/tbsign/tbsign_go /run.sh &&\
    mkdir /app/tbsign/db
ENTRYPOINT ["/run.sh"]

EXPOSE 1323
