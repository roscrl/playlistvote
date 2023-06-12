#!/bin/bash

echo "Helps generate files in this folder like 20220325123456_create_users.sql"
echo "Enter the name of the name for first param 20220325123456_[create]_users.sql"
read first

echo "Enter the name of the name for second param 20220325123456_create_[users].sql"
read second

# Get the current timestamp in the format YYYYMMDDHHMMSS
timestamp=$(date "+%Y%m%d%H%M%S")

filename="$timestamp"_"$first"_"$second".sql

touch "$filename"

echo "Migration file created: $filename"
