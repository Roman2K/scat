# scat

[![godoc][buildbadge]][pipelines] [![godoc][godocbadge]][godoc]

[buildbadge]:https://gitlab.com/Roman2K/scat/badges/release/build.svg
[pipelines]:https://gitlab.com/Roman2K/scat/pipelines
[godocbadge]:https://godoc.org/gitlab.com/Roman2K/scat?status.svg
[godoc]:https://godoc.org/gitlab.com/Roman2K/scat

> Scatter your data away before [loosing][gitlabincident] it

[gitlabincident]:https://about.gitlab.com/2017/02/01/gitlab-dot-com-database-incident/

Backup tool featuring:

* **Decentralization:** avoid trusting any one third-party with all your data

	* data divided into chunks **distributed** anywhere there's space available
	* mix and match **cloud** and local storage in a RAID-like fashion
	* ex: *spread 15GiB of data over 2GiB in Google Drive, 5GiB on a VPS and the rest on an HDD*

* Block-level **de-duplication**

	* [CDC][cdc]-based detection of duplicate blocks, from [restic][restic]
	* **incremental** backups
	* reuse identical blocks of unrelated backups from common remotes
	* immutable storage: stored blocks are never touched upon successive backups
	* ex: *back up a 10GiB sparse disk image with 2GiB used, backup takes <2GiB*
	* ex: *back up VM b, fresh install of the same OS as VM a, backup takes ~MiBs*
	* ex: *append 1 byte to a 1GiB file, next backup takes ~1MiB (last block)*

* RAID-like **error correction**

	* SHA256-based integrity checks ensure data is retrieved unadulterated
	* [Reed-Solomon][b2reedsolomon] erasure coding
	* reconstruct any or all lost shards backed up by any one remote of a given set
	* ex: *I'm locked out of my Google account, recover from Dropbox and Backblaze*
	* ex: *some chunk comes back corrupted from Dropbox, recover from Backblaze and Drive*
	* ex: *my HDD died and I forgot my Backblaze password, recover from Drive and Dropbox*

* **Redundancy:** N-copies duplication, auto-failover on restore

  * Round-Robin spread across eligible remotes
  * increase fault-tolerance of erasure coding
  * ex: *ensure at least 2 copies exist on any two of Drive, Dropbox and some VPS*
  * ex: *back up to 3 cloud providers, recover lost data from up to any 2 of them*

* **Stream**-based: less is more

	* file un-/packing, filtering → tar
	* **snapshot** management → git
	* remote file transfer → ssh
	* **cloud** storage → [rclone](http://rclone.org)
	* asymmetric-key **encryption** → gpg
	* progress, throughput → [pv](http://www.ivarch.com/programs/pv.shtml)
	* Android backup → [Termux](https://termux.com) + ssh

* And:

	* compression
	* multithreaded: configurable concurrency
	* resumable both ways
	* easy to setup, use, and hack on
	* **cross-platform**: binaries for Linux, macOS, Windows, [etc.][builds]

...pick some or all of the above, apply in any order.

Indeed, scat decomposes backing up and restoring into basic stream processors ("procs") arranged like filters in a pipeline. They're chained together, piping the output of proc x to the input of proc x+1. As such, though created for backing up data, its core doesn't actually know anything about backups, but provides the necessary procs.

Such modularity enables unlimited flexibility: stream data from anywhere (local/remote file, arbitrary command, etc.), process it in any way (encrypt, compress, filter through arbitrary command, etc.), to anywhere: upload/download is just another proc at the end/beginning of a chain.

```
                 +---------------------------------+
                 | chain proc                      |
                 |                                 |
+---------+      |  +--------+         +--------+  |
| chunk 0 +----->|  | proc 0 |         | proc 1 |  |
| (seed)  |      |  +--+-----+         +--------+  |
+---------+      |     |                    ^      |
                 |     |    +-------+       |      |
                 |     +--->|+-------+ -----+      |
                 |          +|+-------+            |
                 |           +| chunk |            |
                 |            +-------+            |
                 +---------------------------------+
```

...where `seed` may be a tar stream and procs 0..n would be split, checksum, parity, gzip, scp, etc.

## Setup

1. Download: [latest version][release]
	- versioning scheme: v0, v1, etc.
	- binaries [built][builds] automatically via GitLab's CI
2. Put `scat` in your `$PATH`

## Usage

Stream processing, like performing a backup from a tar stream, is done via a proc chain formulated as a [proc string][procstr]. Below are simple backup-agnostic examples of how to write one (last argument to `scat`).

Hello World:

```sh
$ echo "Hello World" | scat "write[-]"
Hello World

$ echo -n | scat "cmdout[echo Hello World] write[-]"
Hello World

$ echo -n "Hello " | scat "cmd[cat] write[-] cmdout[echo World] write[-]"
Hello World

$ echo "Hello World" | scat "cmd[gpg -e -r 00828C1D] cmd[gpg -d] write[-]"
Hello World

$ echo "Hello World" | scat "cmdin[tee out]" && cat out
Hello World
```

Split `foo`, write chunks to `bar/`:

```sh
$ echo "hello" > foo
$ scat foo "split chain[checksum index[foo_index] cp[bar]]"
$ ls bar
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
```

For restoring, we need a list of all the chunks produced during backup. Proc `index` does that: it lists checksums of chunks output by its containing chain, preserving original order. Note it's part of a subchain following `split`, see [`index`][procindex] for why.

Re-create `foo` from chunk files in `bar/`:

```sh
$ scat foo_index "uindex ucp[bar] uchecksum join[foo]"
$ cat foo
hello
```

The following examples showcase some procs. See [Proc string][procstr] for the full list.

### Example: backup

Example of backing up dir `foo/` to 2 Google Drive accounts and 1 VPS (2 data shards, 1 parity shard, compress, encrypt, checksum, 2 copies, upload - using 8 threads, 4 concurrent transfers)

* seed (stdin): tar stream of `foo/`
* procs: split, compress, parity-split, encrypt, checksum, upload, write index
* output (stdout): index

Command:

```sh
$ tar c foo | scat " \
    split \
    backlog[8 chain[ \
        checksum \
        index[foo_index] \
        gzip \
        parity[2 1] \
        cmd[gpg -e -r 00828C1D] \
        checksum \
	    concur[4 mincopies[2 \
	        [[drive rclone[drive:tmp]] 7gib] \
	        [[drive2 rclone[drive2:tmp]] 14gib] \
	        [bankmon scp[bankmon tmp]] \
	    ]]
	]]"
```

Order matters. Notably:

* split before compressing or encrypting to better detect identical chunks
* compress before parity-split for better ratio
* checksum right after split, before index and at the end, to properly track output chunks: see [`index`][procindex]
* upload within the same chain as `index` so chunks are appended to the index only once successfully uploaded

**Note** both `backlog` and `concur` are being used above. The former limits the number of concurrent instances of `chain` to 8, while the latter limits the number of concurrent uploads by `mincopies` to 4. They may appear redundant, why not one or the other for both? They actually take different types of arguments and have distinct purposes: see [`backlog`][procbacklog] and [`concur`][procconcur].

**Note** the different args in `rclone[drive:tmp]` and `scp[bankmon tmp]`. The former takes a "remote" argument (passed as-is to rclone), while the latter's arguments are "[user@]host" (passed as-is to ssh) and remote directory. See [`rclone`][procrclone] and [`scp`][procscp].

### Example: restore

Reverse chain:

* seed (stdin): index
* procs: read index, download, integrity check, decrypt, parity-join, uncompress, join
* output (stdout): tar stream of `foo`

Command:

```sh
$ scat " \
    uindex \
    backlog[4 multireader[ \
        [drive rclone[drive:tmp]] \
        [drive2 rclone[drive2:tmp]] \
        [bankmon scp[bankmon tmp]] \
    ]]
    backlog[8 chain[ \
        uchecksum \
        cmd[gpg -d] \
        group[3] \
        uparity[2 1] \
		ugzip \
        join[-] \
    ]]" < foo_index | tar x
```

### Snapshots

Making snapshots is as easy as versioning the index file in a git repository:

```sh
$ git init
$ git add foo_index
$ git commit -m "backup of foo"
```

Restoring a snapshot boils down to checking out a particular commit and restoring using the old index file:

```sh
$ git checkout <commit-ish>
$ # ...use foo_index, see Restore
```

You could have a single repository for all your backups and commit index files after each backup.

### Command

```sh
$ scat [options] <proc>
```

Options:

* `-stats` print stats: data rates, quotas, etc.
* `-version` show version
* `-help` show usage

Args:

* `<proc>` proc string, see [Proc string][procstr]

## Rationale

scat is born out of frustration from existing backup solutions:

* [restic][restic], [Borg](https://borgbackup.readthedocs.io), [ZBackup](http://zbackup.org):

  * good: easy to use, block-level deduplication, incremental backups
  * bad: central repository, limited choice of storage engines: local filesystem, SSH, S3

* [git-annex](https://git-annex.branchable.com):

  * good: decentralized, git-based versioning, choice of storage engines (special remotes)
  * bad: difficult to use, file-level deduplication, PITA to compile

* rsync, Drive/Dropbox desktop client:

  * good: easy to use
  * bad: centralized, obscure deduplication if any

* all:

  * bad: reinventing the wheel: each have their own implementation of file un-/packing, pattern-based filtering, snapshot management (save for git-annex), storage engines, encryption, etc.
  * bad: coding style not to my taste, monolythic (if not spaghetti) code base

I wanted to be able to:

* back up anything (one file/dir, some files)
* from anywhere (PC, phone)
* to anywhere (other PC, cloud, vacant space on some VPS)
* when sensible, rely on tools I'm familiar with (ex: tar, git, ssh, rclone, gpg) instead of trusting whether some new tool properly re-implements what existing battle-tested tools already do well

without:

* trusting any third-pary (cloud host, hard drive, VPS host) for reliable storage/retrieval nor privacy
* having to divide at the file-level myself: some dir here, other dir there, that big file doesn't fit anywhere without splitting it
* having to keep track of what's where, let alone copies

I believe scat achieves these objectives 🤓

## Future

I'm very excited to finally have a way to perform backups like I always wanted. I will strive to keep on maintaining it, making sure it stays as simple as possible and fun to hack on.

Should the project be abandoned, existing backups would remain usable with with older versions as well independently of scat using existing tools (shasum, ssh, rclone, gunzip, gpg, cat, etc.).

Upcoming:

* clearer proc string syntax, probably something like piped commands in shells
	* ex:  `foo | bar a b | { baz x | qux y }`
		* instead of `foo bar[a b] chain[baz[x] qux[y]]`
* purge
	* free up space on remotes by garbage-collecting chunks unreachable by given snapshot indexes
		* equivalent of deleting a snapshot in restic and COW filesystems
* streaming file listing
	* lists of existing files are currently buffered due to bad initial decision
		* shouldn't have too much of an impact on memory usage below ~terabytes of data but still feels wrong
* missing unit tests
* logging
* comments for godoc (once the internal API stablized)

## Thanks

* [Gophers](https://golang.org) for such a well thought-out, fun, OCD-satisfying language
* [TJ Holowaychuk](https://twitter.com/tjholowaychuk) for his stunning wielding of simplicity that inspires me everyday
* [Rob Pike](https://twitter.com/rob_pike) for his Go talks, especially [Simplicity is Complicated](https://www.youtube.com/watch?v=rFejpH_tAHM)
* [Klaus Post](https://github.com/klauspost) for [reedsolomon](https://github.com/klauspost/reedsolomon) and his inspirational coding style
* [Alexander Neumann](https://github.com/fd0) for [chunker](https://github.com/restic/chunker)

[restic]:https://github.com/restic/restic
[cdc]:https://restic.github.io/blog/2015-09-12/restic-foundation1-cdc
[b2reedsolomon]:https://www.backblaze.com/blog/reed-solomon

[release]:https://gitlab.com/Roman2K/scat/tags
[builds]:https://gitlab.com/Roman2K/scat/builds
[procstr]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94
[procindex]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#index
[procbacklog]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#backlog
[procconcur]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#concur
[procrclone]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#concur
[procscp]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#concur