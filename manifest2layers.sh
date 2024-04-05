#! /bin/bash

set -e

mkdir -p layers
mkdir -p manifests

flat_name=$(echo "$1" | sed 's/[\/:]/_/g')

crane manifest --platform linux/amd64 "$1" > "manifests/$flat_name"

layers=$( jq -r '.layers.[].digest' < "manifests/$flat_name" )
for layer in $layers; do
    if [ ! -e "layers/$layer" ]; then
        echo "Fetch layer ${layer}"
        crane blob "${1}@${layer}" > "layers/${layer}"
    else
        echo "Cached layer ${layer}"
    fi;
done
