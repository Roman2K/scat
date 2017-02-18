# scat

[![godoc][buildbadge]][pipelines] [![godoc][godocbadge]][godoc]

[buildbadge]:https://gitlab.com/Roman2K/scat/badges/master/build.svg
[pipelines]:https://gitlab.com/Roman2K/scat/pipelines
[godocbadge]:https://godoc.org/gitlab.com/Roman2K/scat?status.svg
[godoc]:https://godoc.org/gitlab.com/Roman2K/scat

> Scatter private data anywhere there's space available

Backup tool that treats its stores as throwaway, untrustworthy commodity

## Features

* **Decentralization:** avoid trusting any one third-party with all your data

	* Round-Robin interleave across uneven storage capacities ~JBOD
	* mesh heterogenous storage hosts: local/remote, big/small, fast/slow
	* automatic redistribution: add/remove stores later
	* ex: *back up 15GiB of data over 2GiB in Google Drive, 5GiB on spare VPS disk space and the rest on my HDD*

* Block-level **de-duplication**

	* [CDC][cdc]-based detection of duplicate blocks, from [restic][restic]
	* **incremental**, immutable backups
	* reuse identical blocks of unrelated backups from common stores
	* ex: *back up a 10GiB pre-allocated disk image with 2GiB used, backup takes <2GiB*
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
	* increase fault-tolerance from erasure coding ~RAID 1+5/6
	* automatic failover on restore
	* ex: *backup in 2 copies among Drive, Backblaze and my HDD*
		* *my HDD died, recover from Drive and B2*
		* *with 1 parity block*
			* *my HDD died and I forgot my Google password, recover from B2*

* **Stream**-based: less is more

	* file un-/packing, filtering → [tar](https://www.gnu.org/software/tar)
	* **snapshot** management → [git](https://git-scm.com)
	* remote file transfer → [ssh](https://www.openssh.com)
	* **cloud** storage → [rclone](http://rclone.org)
	* asymmetric-key **encryption** → [gpg](https://www.gnupg.org)
	* progress, throughput → [pv][pv]
	* Android backup → [Termux](https://termux.com) + ssh

* And:

	* compression
	* multithreaded: configurable concurrency
	* idempotent backup: **resumable**, run often
	* easy to setup, use, and hack on
	* **cross-platform**: binaries for Linux, macOS, Windows, [etc.][builds]

...pick some or all of the above, apply in any order.

Indeed, scat decomposes backing up and restoring into basic stream processors ("procs") arranged like filters in a pipeline. They're chained together, piping the output of proc x to the input of proc x+1. As such, though created for backing up data, its core doesn't actually know anything about backups, but provides the necessary procs.

Such modularity enables unlimited flexibility: stream data from anywhere (local/remote file, arbitrary command, etc.), process it in any way (encrypt, compress, filter through arbitrary command, etc.), to anywhere: write/read/upload/download is just another proc at the end/beginning of a chain.

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

## Demo

[![demo][gif]][video]

Full-length 4K demo video: [on YouTube][video]

[gif]:https://gist.githubusercontent.com/Roman2K/53452c114718d49d25417067334c8955/raw/f2f6f165372015cb08c0f9ce358ef117d9eba20a/out.gif
[video]:https://youtu.be/GcJ8BH4J0WM

## Setup

1. Download: [latest version][release]
	- flat versioning scheme: v0, v1, etc.
	- binaries [built][builds] automatically via GitLab's CI
2. Put `scat` in your `$PATH`

## Usage

Stream processing, like performing a backup from a tar stream, is done via a proc chain formulated as a proc string.

The following example showcase proc strings for typical use cases. They're good starting points to start playing with. Copy them in shell scripts and play around with them, backing up and restoring test files until fully understanding the mechanics at play and reaching desired behaviours. It's important to get comfortable both ways to both back up often and not fear potential moments restoring gets necessary.

See [Proc string][procstr] for syntax documentation and the full list of available procs.

### Backup

Example of backing up dir `foo/` in a RAID 5 fashion to 2 Google Drive accounts and 1 VPS (compress, encrypt, 2 data shards, 1 parity shard, upload >= 2 exclusive copies - using 8 threads, 4 concurrent transfers)

* **seed ← stdin:** tar stream of `foo/`
* **procs:** split, checksum, index track, compress, parity-split, checksum, encrypt, striped upload, index write (implicit)
* **output → stdout:** index

Command:

```bash
$ tar c foo | scat -stats "split | backlog 8 {
  checksum
  | index foo_index
  | gzip
  | parity 2 1
  | checksum
  | cmd gpg --batch -e -r 00828C1D
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
3. `stripe(1 2 ...)`: interleave those across given stores, making `1` copy of each, ensuring at least `2` of 3 are on distinct stores from the others so we can afford to lose any one of them and still be able to recompute original data

Order matters. Notably:

* split before compression and encryption to correctly detect identical chunks
* checksum right after split, before index and after the last producer proc, to properly track output chunks: see [`index`][procindex]
	* but encrypt after final checksum as `gpg -e` is not idempotent, to avoid re-writing/uploading identical chunks
* compress before parity-split and encryption for better ratio
* group before striping: see [`stripe`][procstripe]

> **Note**:
> 
> * Both `backlog` and `concur` are being used above. The former limits the number of concurrent instances of a chain proc (`{}`) to 8, while the latter limits the number of concurrent transfers by `stripe` to 4. They may appear redundant, why not one or the other for both? They actually take different types of arguments and have distinct purposes. See [`backlog`][procbacklog] and [`concur`][procconcur].
> 
> * `rclone(drive:tmp)` and `scp(bankmon tmp)` have a different arguments layout. The former takes a "remote" argument (passed as-is to rclone), while the latter's arguments are "[user@]host" (passed as-is to ssh) and remote directory. See [`rclone`][procrclone] and [`scp`][procscp].

### Restore

Reverse chain:

* **seed ← stdin:** index
* **procs:** index read, download, decrypt, integrity check, parity-join, uncompress, join
* **output → stdout:** tar stream of `foo`

Command:

```bash
$ scat -stats "uindex | backlog 8 {
  backlog 4 multireader(
    drive=rclone(drive:tmp)
    drive2=rclone(drive2:tmp)
    bankmon=scp(bankmon tmp)
  )
  | cmd gpg --batch -d
  | uchecksum
  | group 3
  | uparity 2 1
  | ugzip
  | join -
}" < foo_index | tar x
```

### More

The above only demonstrate a subset of what's possible with scat. There exist more procs and they may be assembled in different manners to tailor to one's particular needs. See [Proc string][procstr].

### Command

```bash
$ scat [options] <proc>
```

Options:

* `-stats` print stats: rates, quotas, etc.
* `-version` show version
* `-help` show usage

Args:

* `<proc>` proc string: see [Proc string][procstr]

### Progress

Being stream-based implies not knowing in advance the total size of data to process. Thus, no progress percentage can be reported. However, when transferring files or directories, size can be known by the caller and passed to [pv][pv].

> **Note:** When piping from pv, do not pass the `-stats` option to scat. Both commands would step on each other's toes writing to stderr and moving terminal cursor.

File backup:

```bash
$ pv my_file | scat "..."
```

Directory backup (approximate progress, not taking into account tar headers):

```bash
# Using GNU du:
$ tar c my_dir | pv -s $(du -sb ~/tmp/100m | cut -f1) | scat "..."

# Under macOS, install GNU coreutils
$ brew install coreutils
$ # idem above, replace du with gdu

# ...or using stock Darwin du, even more approximate:
$ tar c my_dir | pv -s $(du -sk my_dir | cut -f1)k | scat "..."
```

### Snapshots

Making snapshots boils down to versioning the index file in a git repository:

```bash
$ git init
$ git add foo_index
$ git commit -m "backup of foo"
```

Restoring a snapshot consists in checking out a particular commit and restoring using the old index file:

```bash
$ git checkout <commit-ish>
$ # ...use foo_index: see restore example
```

You could have a single repository for all your backups and commit index files after each backup, as well as the backup and restore scripts used to write and read these particular indexes. This allows for modifying proc strings from one backup to the next, while reusing identical chunks if any, and still be able to restore old snapshots created with potentially different proc strings, without having to remember what they were at the time.

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
* when sensible, rely on tools I'm familiar with (ex: tar, git, ssh, rclone, gpg)
	* instead of trusting whether some new tool properly re-implements what existing battle-tested tools already do well

without:

* trusting any third-pary (hard drive, server/cloud provider, etc.) for reliable storage/retrieval nor privacy
* having to divide at the file-level myself: some dir here, other dir there, that big file doesn't fit anywhere without splitting it
* having to keep track of what's where, let alone copies

I believe scat achieves these objectives 🙂

## Next

* See [issues][issueconfirm]
* Subscribe to [Announcements][issueannounce] to get notified about future developments

## Thanks

* [Gophers](https://golang.org) for such a well thought-out, fun, OCD-satisfying language
* [TJ Holowaychuk](https://twitter.com/tjholowaychuk) for his stunning wielding of simplicity that inspires me everyday
* [Rob Pike](https://twitter.com/rob_pike) for his Go talks, especially [Simplicity is Complicated](https://www.youtube.com/watch?v=rFejpH_tAHM)
* [Klaus Post](https://github.com/klauspost) for [reedsolomon](https://github.com/klauspost/reedsolomon) and his inspirational coding style
* [Alexander Neumann](https://github.com/fd0) for [chunker](https://github.com/restic/chunker)

[restic]:https://github.com/restic/restic
[cdc]:https://restic.github.io/blog/2015-09-12/restic-foundation1-cdc
[b2reedsolomon]:https://www.backblaze.com/blog/reed-solomon
[pv]:http://www.ivarch.com/programs/pv.shtml

[release]:https://gitlab.com/Roman2K/scat/tags
[builds]:https://gitlab.com/Roman2K/scat/builds
[issueconfirm]:https://gitlab.com/Roman2K/scat/issues?scope=all&sort=priority&state=opened
[issueannounce]:https://gitlab.com/Roman2K/scat/issues/9
[procstr]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94
[procindex]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#index
[procbacklog]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#backlog
[procconcur]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#concur
[procrclone]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#rclone
[procscp]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#scp
[procstripe]:https://gist.github.com/Roman2K/cc6fd61027306d73f1f2b193f1ce7e94#stripe