#!/bin/bash

# Check if ENV_FILE is provided
if [ -z "$ENV_FILE" ]; then
    echo "ENV_FILE is not set. Using default .env file."
    cp src/.env .env.generated
else
    echo "Using provided ENV_FILE."
    echo "$ENV_FILE" > .env.generated
fi

# Override GO_ENV if provided separately
if [ ! -z "$GO_ENV" ]; then
    sed -i "s/^GO_ENV=.*/GO_ENV=$GO_ENV/" .env.generated
fi

echo "Environment file generated successfully."
