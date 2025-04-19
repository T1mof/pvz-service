#!/bin/sh
# init-db.sh - Скрипт для выполнения миграций из существующей директории

set -e

echo "Проверяем соединение с базой данных..."
until PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c '\q'; do
  >&2 echo "База данных недоступна - ожидаем..."
  sleep 1
done

echo "База данных доступна. Применяем миграции..."

for sql_file in $(find /app/migrations -name "*.sql" | sort); do
  echo "Применяем миграцию: $sql_file"
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $sql_file
  if [ $? -eq 0 ]; then
    echo "Миграция успешно применена: $sql_file"
  else
    echo "Ошибка при применении миграции: $sql_file"
    exit 1
  fi
done

echo "Все миграции успешно применены"