#!/bin/bash

set -e

mkdir -p /data/root
cd /data/root && tar -xvf /out/the_layer.tar
cd /out
make_ext4fs -l 1G -b 64k -L stream -g 256 root.img
mkdir -p /tmp/disk
fuse2fs  root.img /tmp/disk -o rw
cp -av /data/root/* /tmp/disk/
umount /tmp/disk
