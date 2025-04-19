#!/bin/sh
# wait-for-it.sh - Простой скрипт для ожидания доступности порта

set -e

host="$1"
port="$2"
shift 2
cmd="$@"

until nc -z "$host" "$port"; do
  >&2 echo "Ожидаем доступности $host:$port..."
  sleep 1
done

>&2 echo "$host:$port доступен"

if [ -n "$cmd" ]; then
  >&2 echo "Выполняем команду: $cmd"
  exec $cmd
fi