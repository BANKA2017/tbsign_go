FROM alpine:3 AS build

WORKDIR /app/tbsign

ARG TARGETARCH
COPY tbsign_go.musl.${TARGETARCH}.docker/tbsign_go /app/tbsign/

RUN chmod +x /app/tbsign/tbsign_go /app/tbsign/run.sh &&\
    mkdir /app/tbsign/db

FROM alpine:3
COPY --from=build /app/tbsign/ /app/tbsign/

ENV tc_db_path="/app/tbsign/db/tbsign_go.db" \
    tc_api="true" \
    tc_fe="true" \
    tc_address=":1323"

ENTRYPOINT ["/app/tbsign/tbsign_go"]

EXPOSE 1323
