#! /bin/bash
#
# Depends on :
# - crane from https://github.com/google/go-containerregistry
# - jq
#
# Usage example :
# $ ./manifest2layers.sh debian:12-slim
#
# The script fetch the manifest for linux/amd64, et and lazy get the layers.
# Datas are stored in layers and manifests folders.
#
set -e

mkdir -p layers
mkdir -p manifests

# / and : are boring in file name, lets replace them with _
flat_name=$(echo "$1" | sed 's/[\/:]/_/g')
manifest="manifests/$flat_name"

crane manifest --platform linux/amd64 "$1" > $manifest

layers=$( jq -r '.layers.[].digest' < $manifest )
for layer in $layers; do
    if [ ! -e "layers/$layer" ]; then
        echo "Fetch layer ${layer}"
        crane blob "${1}@${layer}" > "layers/${layer}"
    else
        echo "Cached layer ${layer}"
    fi;
done
