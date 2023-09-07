#!/bin/sh
redis-server &
while ! redis-cli -u redis://localhost:6379 ping; do
  sleep 1s
done
/puremoot
