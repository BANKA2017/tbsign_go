FROM alpine:3

WORKDIR /app/tbsign

ARG TARGETARCH
COPY tbsign_go.musl.${TARGETARCH}/tbsign_go /app/tbsign/

ENV tc_db_path="/app/tbsign/tbsign_go.db" \
    tc_api="true" \
    tc_fe="true"

CMD ["sh", "-c", "chmod +x tbsign_go && ./tbsign_go"]

EXPOSE 8080
