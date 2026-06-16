#!/bin/bash


echo "running migrations"

chmod +x ./wait-for-it.sh

./wait-for-it.sh $DATABASE_HOST:$DATABASE_PORT --timeout=5 -- goose -dir=/migrations status
./wait-for-it.sh $DATABASE_HOST:$DATABASE_PORT --timeout=5 -- goose -dir=/migrations up

echo "migrations done"