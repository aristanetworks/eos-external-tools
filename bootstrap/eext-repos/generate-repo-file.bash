#!/bin/bash

set -e
set -x

usage() {
   echo "Usage: $0 <template_file_path> <environment_file_path> <output_file_path>"
   exit 1
}

if [ $# -ne 3 ]; then
   usage
fi

TEMPLATE_FILE=$1
ENV_FILE=$2
OUTPUT_FILE=$3

if [ ! -f "$TEMPLATE_FILE" ]; then
   echo "Error: Template file '$TEMPLATE_FILE' not found."
   exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
   echo "Error: Environment file '$ENV_FILE' not found."
   exit 1
fi

# Load environment variables from the environment file
set -a
source "$ENV_FILE"
set +a

envsubst < "$TEMPLATE_FILE" > "$OUTPUT_FILE"

