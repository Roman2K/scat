# scat

[![godoc][buildbadge]][pipelines] [![godoc][godocbadge]][godoc]

[buildbadge]:https://gitlab.com/Roman2K/scat/badges/master/build.svg
[pipelines]:https://gitlab.com/Roman2K/scat/pipelines
[godocbadge]:https://godoc.org/gitlab.com/Roman2K/scat?status.svg
[godoc]:https://godoc.org/gitlab.com/Roman2K/scat

> Scatter your data away [before][codinghorrorincident] [loosing][gitlabincident] [it][githubincident]

[codinghorrorincident]:https://blog.codinghorror.com/international-backup-awareness-day/
[gitlabincident]:https://about.gitlab.com/2017/02/01/gitlab-dot-com-database-incident/
[githubincident]:https://github.com/blog/744-today-s-outage

Backup tool that treats storage hosts as throwaway, untrustworthy commodity

## Features

* **Decentralization:** avoid trusting any one third-party with all your data

	* Round-Robin interleave across uneven storage capacities ~JBOD
	* mesh heterogenous storage hosts: local/remote, big/small, fast/slow
	* automatic redistribution: add/remove hosts later
	* ex: *back up 15GiB of data over 2GiB in Google Drive, 5GiB on spare VPS disk space and the rest on my HDD*

* Block-level **de-duplication**

	* [CDC][cdc]-based detection of duplicate blocks, from [restic][restic]
	* **incremental**, immutable backups
	* reuse identical blocks of unrelated backups from common hosts
	* ex: *back up a 10GiB sparse disk image with 2GiB used, backup takes <2GiB*
	* ex: *append 1 byte to a 1GiB file, next backup takes ~1MiB (last block)*

* RAID-like **error correction**

	* striping with distributed parity ~RAID 5/6
	* grow/shrink array later
	* SHA256-based integrity checks
	* [Reed-Solomon][b2reedsolomon] erasure coding
	* ex: *backup with 1 parity block among Drive, Backblaze and my HDD*
		* *some block comes back corrupted from my HDD, recover from Drive and B2*
		* *I'm locked out of my Google account, recover from B2 and my HDD*

* **Redundancy:** N-copies duplication

	* ensure N+ copies exist at all times ~RAID 1
	* automatic failover on restore
	* increase fault-tolerance from erasure coding ~RAID 1+5/6
	* ex: *backup in 2 copies among Drive, Backblaze and my HDD*
		* *my HDD died, recover from Drive and B2*
		* *with 1 parity block*
			* *my HDD died and I forgot my Google password, recover from B2*

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
	* idempotent backup: **resumable**, run often
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

...where `seed` may be a tar stream and procs 0..n would be split, checksum, parity, gzip, scp, etc. part of a chain that is itself a proc also.

## Setup

1. Download: [latest version][release]
	- versioning scheme: v0, v1, etc.
	- binaries [built][builds] automatically via GitLab's CI
2. Put `scat` in your `$PATH`

## Usage

Stream processing, like performing a backup from a tar stream, is done via a proc chain formulated as a [proc string][procstr]. Below are simple backup-agnostic examples of how to write one (last argument to `scat`).

Hello World:

```sh
$ echo "Hello World" | scat "write -"
Hello World

$ scat "cmdout echo Hello World | write -" < /dev/null
Hello World

$ echo -n "Hello " | scat "cmd cat | write - | cmdout echo World | write -"
Hello World

$ echo "Hello World" | scat "cmd gpg -e -r 00828C1D | cmd gpg -d --batch | write -"
Hello World

$ echo "Hello World" | scat "cmdin tee hello" && cat hello
Hello World
```

Split file `foo`, write chunks to `bar/`:

```sh
$ echo hello > foo
$ scat "split | { checksum | index - | cp bar }" < foo > foo_index
$ ls bar
5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03
```

For restoring, we need a list of all the chunks produced during backup. Proc `index` does that: it lists checksums of chunks output by its containing chain, preserving original order. **Note** it's part of a subchain (`{}`), following `split`: see [`index`][procindex].

Re-create `foo` from chunk files in `bar/`:

```sh
$ scat "uindex | ucp bar | uchecksum | join -" < foo_index > foo
$ cat foo
hello
```

The above are just generic examples to get familiar with the proc string. As to how to actually back up and restore, one would specify such a proc string employing the procs tailored to their particular needs. See [Proc string][procstr] for the full list.

The following examples are good starting points for typical needs. Copy them in shell scripts and play around with them, backing up and restoring test files until fully understanding the mechanics under the hood and reaching desired behaviours. It's important to get comfortable both ways to both back up often and not fear potential moments restoring gets necessary.

### Example: backup

Example of backing up dir `foo/` in a RAID 5 fashion to 2 Google Drive accounts and 1 VPS (compress, encrypt, 2 data shards, 1 parity shard, upload >= 2 distinct copies - using 8 threads, 4 concurrent transfers)

* **seed ← stdin:** tar stream of `foo/`
* **procs:** split, checksum, index track, compress, parity-split, checksum, encrypt, striped upload, index write (implicit)
* **output → stdout:** index

Command:

```sh
$ tar c foo | scat -stats "split | backlog 8 {
  checksum
  | index foo_index
  | gzip
  | parity 2 1
  | checksum
  | cmd gpg -e -r 00828C1D
  | group 3
  | concur 4 stripe(1 2
      mydrive=rclone(drive:tmp)=7gib
      mydrive2=rclone(drive2:tmp)=14gib
      myvps=scp(bankmon tmp)
    )
  }"
```

The combination of `parity`, `group` and `stripe` creates a RAID 5:

1. `parity(2 1)`: split into `2` data shards and `1` parity shard
2. `group(3)`: aggregate all `3` shards for striping
3. `stripe(1 2 ...)`: interleave those across given hosts, making `1` copy of each, ensuring at least `2` of 3 are on distinct hosts from the others so we can afford to lose any one of them

**Note** that order matters. Notably:

* split before compression and encryption to correctly detect identical chunks
* checksum right after split, before index and after the last producer proc, to properly track output chunks: see [`index`][procindex]
	* but encrypt after final checksum as `gpg -e` is not idempotent, to avoid re-uploading identical chunks
* compress before parity-split and encryption for better ratio
* group before striping: see [`stripe`][procstripe]

**Note** both `backlog` and `concur` are being used above. The former limits the number of concurrent instances of a chain proc (`{}`) to 8, while the latter limits the number of concurrent transfers by `stripe` to 4. They may appear redundant, why not one or the other for both? They actually take different types of arguments and have distinct purposes: see [`backlog`][procbacklog] and [`concur`][procconcur].

**Note** the different args in `rclone(drive:tmp)` and `scp(bankmon tmp)`. The former takes a "remote" argument (passed as-is to rclone), while the latter's arguments are "[user@]host" (passed as-is to ssh) and remote directory. See [`rclone`][procrclone] and [`scp`][procscp].

### Example: restore

Reverse chain:

* **seed ← stdin:** index
* **procs:** index read, download, decrypt, integrity check, parity-join, uncompress, join
* **output → stdout:** tar stream of `foo`

Command:

```sh
$ scat -stats "uindex | backlog 8 {
  backlog 4 multireader(
    drive=rclone(drive:tmp)
    drive2=rclone(drive2:tmp)
    bankmon=scp(bankmon tmp)
  )
  | cmd gpg -d --batch
  | uchecksum
  | group 3
  | uparity 2 1
  | ugzip
  | join -
}" < foo_index | tar x
```

### Snapshots

Making snapshots boils down to versioning the index file in a git repository:

```sh
$ git init
$ git add foo_index
$ git commit -m "backup of foo"
```

Restoring a snapshot consists in checking out a particular commit and restoring using the old index file:

```sh
$ git checkout <commit-ish>
$ # ...use foo_index: see restore example
```

You could have a single repository for all your backups and commit index files after each backup, as well as the backup and restore scripts used to write and read these particular indexes. This allows for modifying proc strings from one backup to the next, while reusing identical chunks if any, and still be able to restore an old snapshot created with a potentially different proc string, without having to remember what it was at the time.

### Command

```sh
$ scat [options] <proc>
```

Options:

* `-stats` print stats: rates, quotas, etc.
* `-version` show version
* `-help` show usage

Args:

* `<proc>` proc string: see [Proc string][procstr]

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

* trusting any third-pary (hard drive, server/cloud provider, etc.) for reliable storage/retrieval nor privacy
* having to divide at the file-level myself: some dir here, other dir there, that big file doesn't fit anywhere without splitting it
* having to keep track of what's where, let alone copies

I believe scat achieves these objectives 🤓

## Upcoming

* purge
	* free up space on hosts by garbage-collecting chunks unreachable by given snapshot indexes
		* equivalent of deleting a snapshot in restic and COW filesystems
* code cleanups
	* streaming file listing
		* lists of existing files are currently buffered due to bad initial decision
			* shouldn't have too much of an impact on memory usage below ~terabytes of data but still feels wrong
	* finer grained quota filling for exclusive striping
		* currently, if chunks are grouped before striping, the total size is used to determine if there's space available, not the size of each chunk individually
* logging
* missing unit tests
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
[procrclone]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#rclone
[procscp]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#scp
[procstripe]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#stripe