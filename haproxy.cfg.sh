#!/usr/bin/env bash

echo "
defaults
        mode    tcp
        timeout client      30000ms
        timeout server      30000ms
        timeout connect      3000ms

frontend fr_server1
        bind 0.0.0.0:80
        stick-table type ip size 100k expire 30s store conn_rate(1s)
        tcp-request connection track-sc0 src
        tcp-request connection reject if { sc0_conn_rate gt ${RATE_LIMIT} }
        default_backend bk_server1

backend bk_server1
        server srv1 host.docker.internal:8080 maxconn 2048
"
