FROM busybox:1.37-uclibc AS build

WORKDIR /app/tbsign

ARG TARGETARCH
COPY docker/tbsign_go.linux-${TARGETARCH} /app/tbsign/tbsign_go

RUN chmod +x /app/tbsign/tbsign_go &&\
    mkdir /app/tbsign/db

FROM busybox:1.37-uclibc

COPY --from=build /app/tbsign/ /app/tbsign/

ENV tc_db_path="/app/tbsign/db/tbsign_go.db" \
    tc_api="true" \
    tc_fe="true" \
    tc_address=":1323" \
    tc_docker_mode="true"

ENTRYPOINT ["/app/tbsign/tbsign_go"]

EXPOSE 1323
