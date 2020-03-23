#!/usr/bin/env bash

function setup() {
    if [[ -z "${RATE_LIMIT}" ]]; then
      echo "ensure RATE_LIMIT envvar is set (in terms of req/seconds)."
      return false
    fi
    ./haproxy.cfg.sh > haproxy.cfg

    docker rm -f concurrent-http-client-haproxy concurrent-http-client 2>&1 1>/dev/null || true
    docker run --name concurrent-http-client-haproxy -d --rm -p 8081:80 -v $PWD/haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg haproxy:1.7 > /dev/null
    docker run --name concurrent-http-client -d --rm -p 8080:8080 -v $PWD:/work -w /work golang:1.13-stretch sleep infinity > /dev/null
    docker exec -it concurrent-http-client go mod download > /dev/null
    echo "All set."
}

function run_server() {
    docker exec -it concurrent-http-client go run ./cmd/server "$@"
}

function run_client() {
    docker exec -it concurrent-http-client go run ./cmd/client "$@"
}
