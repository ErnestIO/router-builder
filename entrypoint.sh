#!/usr/bin/env sh

echo "Waiting for NATS"
while ! echo exit | nc nats 4222; do sleep 1; done

echo "Starting router-builder"
/go/bin/router-builder
