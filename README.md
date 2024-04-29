# Stream my root

AWS wrote a white paper about Lambda, and explain to stream container images, for a better startup time.
Lets play with this tool, containers (even tiny VM) are always fun to manipulate.

[On-demand Container Loading in AWS Lambda](https://arxiv.org/abs/2305.13162)

* [x] Build desterministic blocks from oci/docker images
* [x] Expose chunks as nbd
* [ ] Expose chunks without nbd
* [ ] Image is writable, with a COW and a map
* [ ] Chunks are crypted, recipe has keys
* [ ] Don't compress, just trimme zeros
* [ ] Lazy download chunks with HTTP
* [ ] Build the list of needed chunks to start an image
* [ ] Garbage collect unused chunks

## Use it

Build the make_ext4fs image

```bash
make make_ext4fs
```

Build the tool image

```bash
make docker-tool
```

Fetch some images

```bash
make img NAME=gcr.io/distroless/python3-debian12
make img NAME=gcr.io/distroless/base-debian12
# *.img files are stored in out/
ls out
# images ar full of holes
ls -lsh out/*.img
# holes are here
filefrag -v out/gcr.io_distroless_base-debian12.img
```

Disk images are well handled by file

```bash
$ file out/gcr.io_distroless_python3-debian12.img
out/gcr.io_distroless_python3-debian12.img: Linux rev 1.0 ext4 filesystem data, UUID=d1fa2f31-4aeb-8354-9262-b4d19504856c, volume name "stream" (extents) (large files)
```

Build tools (with golang)

```bash
make
```

Cut images in small chunks

```bash
./bin/chunk out/*.img
# recipe are stored near the img file
ls out
# chunks are stored is smr/
ls smr
```

Run the server

```bash
# the first arg is a recipe
./bin/server out/gcr.io_distroless_python3-debian12.img.recipe
```

Qemu can see it

```bash
$ qemu-img info nbd://localhost:10809/smr
image: nbd://localhost:10809/smr
file format: raw
virtual size: 1 GiB (1073741824 bytes)
disk size: unavailable
Child node '/file':
    filename: nbd://localhost:10809/smr
    protocol type: nbd
    file length: 1 GiB (1073741824 bytes)
    disk size: unavailable
```

Mount the image (from a Linux)

```bash
# split your tmux with ctrl-% and watch the kernel yelling at nbd
tail -f /var/log/kern.log
# nbd module should be loaded
sudo modprobe nbd
sudo nbd-client -N smr localhost 10809 /dev/nbd1
sudo mkdir /mnt/smr
sudo mount -o ro -t ext4 /dev/nbd1 /mnt/smr
ls /mnt/smr
```

Mount from a VM on a Mac

It works with

* [Lima](https://lima-vm.io) `brew install lima`. The image is minimalist, without kernel logging making debug a bit harder.
* [Multipass](https://multipass.run/) `brew install multipass`. A good old fat Ubuntu image.
* Vagrant. Old, huge, but it works.

Don't use `localhost` but the host IP.

## Test it

Some fixtures

```bash
make img NAME=gcr.io/distroless/base-debian12
./bin/chunk out/gcr.io_distroless_base-debian12.img
```

Test (and even fuzzing)

```bash
make test
make fuzz
```

Compare chunked and plain images

```bash
./bin/debug out/gcr.io_distroless_base-debian12.img
```

## Centralized image chunks

In real world, chunks should be stored in a replicated/cenralized place, some S3 clones, or even a Registry.

In the paper, each container has its sidekick service handling its "lazy disk", but chunks are available for all sidekicks on a server.

Each servers share a local cache, with coordinated downloading : a chunk should be downloaded once, even if it is required for starting multiple containers.

Cached just handles old chunks evictions and locking for downloading a distant chunk.

```
        +-------------+
        | Chunk store |
        +------+------+
               |
+--------------|-----------+
| +--------+   |           |
| | Cached |   |           |
| +---+----+   |           |
|     |        |           |
| +---+--------+--+        |
| |    Storaged   |        |
| +---------------+        |
|                          |
+--------------------------+
```

### Cached

Cached manages the chunks lifecycle with a LRU-K, just like a plain old spinning disk.

Cached exposes some rpc via a minimalistic protocol ([yamux](https://pkg.go.dev/github.com/hashicorp/yamux) over UNIX socket), without serialisation (binary in, binary out), msg size are announced, no buffer needed.

Locking is simple : the first user to ask for a key get a "true" answer, and try to fetch the file to the download file, then release the lock with another rpc call.
Other users asking for the same key get no answer in first time, just wait the release of the lock,
then get a "false" answer.

The paper says that Lambda use one more level of cache, some key/value service.

## Stuff to read

* https://github.com/opencontainers/image-spec/blob/main/layer.md
* https://reproducible-builds.org/docs/system-images/
