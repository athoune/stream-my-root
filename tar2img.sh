#!/bin/bash

set -e

# / and : are boring in file name, lets replace them with _
flat_name=$(echo "$1" | sed 's/[\/:]/_/g')
manifest="manifests/$flat_name"
layers=$( jq -r '.layers.[].digest' < "$manifest" )
image=out/${flat_name}.img
cwd=$(pwd)

make_ext4fs -l 1G -b 64k -L stream -g 256 "$image"

mkdir -p /tmp/disk
fuse2fs  "$image" /tmp/disk -o rw
cd /tmp/disk

for layer in $layers; do
    # FIXME .wh. name must be handled
    tar -xvzf "$cwd/layers/$layer"
done

cd "$cwd"
umount /tmp/disk
