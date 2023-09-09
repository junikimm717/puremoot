#!/bin/bash

redis-server --appendonly yes &
while ! redis-cli -u redis://localhost:6379 ping; do
  sleep 1s
done
exec /puremoot
