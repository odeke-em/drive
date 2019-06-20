# drive

[![Build Status](https://travis-ci.org/odeke-em/drive.png?branch=master)](https://travis-ci.org/odeke-em/drive)

`drive` is a tiny program to pull or push [Google Drive](https://drive.google.com) files.

`drive` was originally developed by [Burcu Dogan](https://github.com/rakyll) while working on the Google Drive team. Since she is very busy and no longer able to maintain it, I took over drive on `Thursday, 1st January 2015`. This repository contains the latest version of the code.

## Table of Contents

- [Installing](#installing)
  - [Requirements](#requirements)
  - [From sources](#from-sources)
  - [Godep](#godep)
  - [Platform Packages](#platform-packages)
    - [Automation Scripts](#automation-scripts)
  - [Cross Compilation](#cross-compilation)
  - [API keys](#api-keys)
- [Usage](#usage)
  - [Hyphens: - vs --](#hyphens---vs---)
  - [Initializing](#initializing)
  - [De Initializing](#de-initializing)
  - [Traversal Depth](#traversal-depth)
  - [Configuring General Settings](#configuring-general-settings)
  - [Excluding And Including Objects](#excluding-and-including-objects)
    - [Sample .driveignore with the exclude and include clauses combined](sample-.driveignore-with-the-exclude-and-include-clauses-combined)
  - [Pulling](#pulling)
    - [Verifying Checksums](#verifying-checksums)
    - [Exporting Docs](#exporting-docs)
  - [Pushing](#pushing)
  - [Pulling And Pushing Notes](#pulling-and-pushing-notes)
  - [End to End Encryption](#end-to-end-encryption)
  - [Publishing](#publishing)
  - [Unpublishing](#unpublishing)
  - [Sharing and Emailing](#sharing-and-emailing)
  - [Unsharing](#unsharing)
  - [Starring Or Unstarring](#starring-or-unstarring)
  - [Diffing](#diffing)
  - [Touching](#touching)
  - [Trashing And Untrashing](#trashing-and-untrashing)
  - [Emptying The Trash](#emptying-the-trash)
  - [Deleting](#deleting)
  - [Listing](#listing)
  - [Stating](#stating)
  - [Printing URL](#printing-url)
  - [Editing Description](#editing-description)
  - [Retrieving MD5 Checksums](#retrieving-md5-checksums)
  - [Retrieving FileId](#retrieving-fileid)
  - [Retrieving Quota](#retrieving-quota)
  - [Retrieving Features](#retrieving-features)
  - [Creating](#creating)
  - [Opening](#opening)
  - [Copying](#copying)
  - [Moving](#moving)
  - [Renaming](#renaming)
  - [Command Aliases](#command-aliases)
  - [Detecting And Fixing Clashes](#detecting-and-fixing-clashes)
  - [.desktop Files](#desktop-files)
  - [Fetching And Pruning Missing Index Files](#fetching-and-pruning-missing-index-files)
  - [Drive Server](#drive-server)
  - [QR Code Share](#qr-code-share)
  - [About](#about)
  - [Help](#help)
  - [Filing Issues](#filing-issues)
- [Revoking Account Access](#revoking-account-access)
- [Uninstalling](#uninstalling)
- [Applying Patches](#applying-patches)
- [Why Another Google Drive Client?](#why-another-google-drive-client)
- [Known Issues](#known-issues)
- [Reaching Out](#reaching-out)
- [Disclaimer](#disclaimer)
- [LICENSE](#license)

## Installing

### Requirements

go 1.9.X or higher is required. See [here](https://golang.org/doc/install) for installation instructions and platform installers.

* Make sure to set your GOPATH in your env, .bashrc or .bash\_profile file. If you have not yet set it, you can do so like this:

```shell
cat << ! >> ~/.bashrc
> export GOPATH=\$HOME/gopath
> export PATH=\$GOPATH:\$GOPATH/bin:\$PATH
> !
source ~/.bashrc # To reload the settings and get the newly set ones # Or open a fresh terminal
```
The above setup will ensure that the drive binary after compilation can be invoked from your current path.

### From sources

To install from the latest source, run:

```shell
go get -u github.com/odeke-em/drive/cmd/drive
```

Otherwise:

* In order to address [issue #138](https://github.com/odeke-em/drive/issues/138), where debug information should be bundled with the binary, you'll need to run:

```shell
go get github.com/odeke-em/drive/drive-gen && drive-gen
```

In case you need a specific binary e.g for Debian folks [issue #271](https://github.com/odeke-em/drive/issues/271) and [issue 277](https://github.com/odeke-em/drive/issues/277)

```shell
go get -u github.com/odeke-em/drive/drive-google
```

That should produce a binary `drive-google`

OR

To bundle debug information with the binary, you can run:

```shell
go get -u github.com/odeke-em/drive/drive-gen && drive-gen drive-google
```

### Godep

+ Using godep
```
cd $GOPATH/src/github.com/odeke-em/drive/drive-gen && godep save
```

+ Unravelling/Restoring dependencies
```
cd $GOPATH/src/github.com/odeke-em/drive/drive-gen && godep restore
```

Please see file `drive-gen/README.md` for more information.

### Platform Packages

For packages on your favorite platform, please see file [Platform Packages.md](https://github.com/odeke-em/drive/blob/master/platform_packages.md).

Is your platform missing a package? Feel free to prepare / contribute an installation package and then submit a PR to add it in.

#### Automation Scripts

You can install scripts for automating major drive commands and syncing from [drive-google wiki](https://gitlab.com/jean-christophe-manciot/Drive/wikis/Drive-Automation-Scripts:-Startup-Guide-and-Screenshots), also described in [platform_packages.md](https://github.com/odeke-em/drive/blob/master/platform_packages.md).
Some screenshots are available [here](https://gitlab.com/jean-christophe-manciot/Drive/wikis/Drive-Automation-Scripts:-Startup-Guide-and-Screenshots#drive-menu-screenshots).

### Cross Compilation

See file `Makefile` which currently supports cross compilation. Just run `make` and then inspect the binaries in directory `bin`.

* Supported platforms to cross compile to:
- ARMv5.
- ARMv6.
- ARMv7.
- ARMv8.
- Darwin (OS X).
- Linux.

Also inspect file `bin/md5Sums.txt` after the cross compilation.

### API keys

Optionally set the `GOOGLE_API_CLIENT_ID` and `GOOGLE_API_CLIENT_SECRET` environment variables to use your own API keys.

## Usage

### Hyphens: - vs --

A single hyphen `-` can be used to specify options. However two hyphens `--` can be used with any options in the provided examples below.

### Initializing

Before you can use `drive`, you'll need to mount your Google Drive directory on your local file system:

#### OAuth2.0 credentials
```shell
drive init ~/gdrive
cd ~/gdrive
```

#### Google Service Account credentials
```shell
drive init --service-account-file <gsa_json_file_path> ~/gdrive
cd ~/gdrive
```

Where <gsa_json_file_path> must be a [Google Service Account credentials](https://developers.google.com/identity/protocols/OAuth2ServiceAccount#creatinganaccount) file in JSON form.
This feature was implemented as requested by:
+ https://github.com/odeke-em/drive/issues/879


### De Initializing

The opposite of `drive init`, it will remove your credentials locally as well as configuration associated files.

```shell
drive deinit [-no-prompt]
```

For a complete deinitialization, don't forget to revoke account access, [please see revoking account access](#revoking-account-access)

### Traversal Depth

Before talking about the features of drive, it is useful to know about "Traversal Depth".

Throughout this README the usage of the term "Traversal Depth" refers to the number of

nodes/hops/items that it takes to get from one parent to children. In the options that allow it, you'll have a flag option `-depth <n>` where n is an integer

* Traversal terminates on encountering a zero `0` traversal depth.

* A negative depth indicates infinity, so traverse as deep as you can.

* A positive depth helps control the reach.

Given:

	|- A/
		|- B/
		|- C/
			|- C1
			|- C2
				|- C10/
				|- CTX/
					| - Music
					| - Summary.txt

+ Items on the first level relative to A/ ie `depth 1`, we'll have:

  B, C

+ On the third level relative to C/ ie `depth 3`

  * We'll have:
  
    Items: Music, Summary.txt

  * The items encountered in `depth 3` traversal relative to C/ are:

				|- C1
				|- C2
					|- C10/
					|- CTX/
						| - Music
						| - Summary.txt

+ No items are within the reach of  `depth -1` relative to B/ since B/ has no children.

+ Items within the reach of `depth -` relative to CTX/ are:

				| - Music
				| - Summary.txt

### Configuring General Settings

drive supports resource configuration files (.driverc) that you can place both globally (in your home directory)
and locally(in the mounted drive dir) or in the directory that you are running an operation from, relative to the root.
The entries for a .driverc file is in the form a key-value pair where the key is any of the arguments that you'd get
from running
```shell
drive <command> -h
# e.g
drive push -h
```

and the value is the argument that you'd ordinarily supply on the commandline.
.driverc configurations can be optionally grouped in sections. See https://github.com/odeke-em/drive/issues/778.

For example:

```shell
cat << ! >> ~/.driverc
> # My global .driverc file
> export=doc,pdf
> depth=100
> no-prompt=true
>
> # For lists
> [list]
> depth=2
> long=true
>
> # For pushes
> [push]
> verbose=false
>
> # For stats
> [stat]
> depth=3
>
> # For pulls and pushes
> [pull/push]
> no-clobber=true
> !

cat << ! >> ~/emm.odeke-drive/.driverc
> # The root main .driverc
> depth=-1
> hidden=false
> no-clobber=true
> exports-dir=$HOME/exports
> !

cat << $ >> ~/emm.odeke-drive/fall2015Classes/.driverc
> # My global .driverc file
> exports-dir=$HOME/Desktop/exports
> export=pdf,csv,txt
> hidden=true
> depth=10
> exclude-ops=delete,update
> $
```

### Excluding and Including Objects

drive allows you to specify a '.driveignore' file similar to your .gitignore, in the root
directory of the mounted drive. Blank lines and those prefixed by '#' are considered as comments and skipped.

For example:

```shell
cat << $ >> .driveignore
> # My drive ignore file
> \.gd$
> \.so$
> \.swp$
> $
```

Note:
  * Pattern matching and suffixes are done by regular expression matching so make sure to use a valid regular expression suffix.

  * Go doesn't have a negative lookahead mechanism ie `exclude all but` which would
    normally be achieved in other languages or regex engines by "?!". See https://groups.google.com/forum/#!topic/golang-nuts/7qgSDWPIh_E.
    This was reported and requested in [issue #535](https://github.com/odeke-em/drive/issues/535).
    A use case might be ignoring all but say .bashrc files or .dotfiles.
    To enable this, prefix "!" at the beginning of the path to achieve this behavior.

#### Sample .driveignore with the exclude and include clauses combined
```shell
cat << $ >> .driveignore
> ^\.
> !^\.bashrc # .bashrc files won't be ignored
> _export$ # _export files are to be ignored
> !must_export$ # the exception to the clause anything with "must_export"$ won't be ignored
```

### Pulling

The `pull` command downloads data that does not exist locally but does remotely on Google drive, and may delete local data that is not present on Google Drive. 
Run it without any arguments to pull all of the files from the current path:

```shell
drive pull
```

To pull and decrypt your data that is stored encrypted at rest on Google Drive, use flag `-decryption-password`:

See [Issue #543](https://github.com/odeke-em/drive/issues/543)

```shell
drive pull -decryption-password '$JiME5Umf' influx.txt
```

Pulling by matches is also supported

```shell
cd ~/myDrive/content/2015
drive pull -matches vines docx
```

To force download from paths that otherwise would be marked with no-changes

```shell
drive pull -force
```

To pull specific files or directories, pass in one or more paths:

```shell
drive pull photos/img001.png docs
```

Pulling by id is also supported

```shell
drive pull -id 0fM9rt0Yc9RTPaDdsNzg1dXVjM0E 0fM9rt0Yc9RTPaTVGc1pzODN1NjQ 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU
```

`pull` optionally allows you to pull content up to a desired depth.

Say you would like to get just folder items until the second level

```shell
drive pull -depth 2 heavy-files summaries
```

Traverse deep to infinity and beyond

```shell
drive pull -depth -1 all-my-files
```

Pulling starred files is allowed as well

```shell
drive pull -starred
drive pull -starred -matches content
drive pull -starred -all # Pull all the starred files that aren't in the trash
drive pull -starred -all -trashed # Pull all the starred files in the trash
```

Like most commands [.driveignore](#excluding-and-including-objects) can be used to filter which files to pull.

+ Note: Use `drive pull -hidden` to also pull files starting with `.` like `.git`.

To selectively pull by type e.g file vs directory/folder, you can use flags
- `files`
- `directories`

```shell
drive pull -files a1/b2
drive pull -directories tf1
```

#### Verifying Checksums
Due to popular demand, by default, checksum verification is turned off. It was deemed to be quite vigorous and unnecessary for most cases, in which size + modTime differences are sufficient to detect file changes. The discussion stemmed from issue [#117](https://github.com/odeke-em/drive/issues/117).

However, modTime differences on their own do not warrant a resync of the contents of file.
Modification time changes are operations of their own and can be made:
+ locally by, touching a file (chtimes).
+ remotely by just changing the modTime meta data.

To turn checksum verification back on:

```shell
drive pull -ignore-checksum=false
```

drive also supports piping pulled content to stdout which can be accomplished by:

```shell
drive pull -piped path1 path2
```

+ In relation to issue #529, you can change the max retry counts for exponential backoff. Using a count < 0 falls back to the
default count of 20:
```shell
drive pull -retry-count 14 documents/2016/March videos/2013/September
```

#### Exporting Docs

By default, the `pull` command will export Google Docs documents as PDF files. To specify other formats, use the `-export` option:

```shell
drive pull -export pdf,rtf,docx,txt
```

To explicitly export instead of using `-force`

```shell
drive pull -export pdf,rtf,docx,txt -explicitly-export
```

By default, the exported files will be placed in a new directory suffixed by `\_exports` in the same path. To export the files to a different directory, use the `-exports-dir` option:

```shell
drive pull -export pdf,rtf,docx,txt -exports-dir ~/Desktop/exports
```

Otherwise, you can export files to the same directory as requested in [issue #660](https://github.com/odeke-em/drive/issues/660),
by using pull flag `-same-exports-dir`. For example:
```shell
drive pull -explicitly-export -exports-dir ~/Desktop/exp -export pdf,txt,odt -same-exports-dir 
Resolving...
+ /test-exports/few.docs
+ /test-exports/few
+ /test-exports/influx
Addition count 3
Proceed with the changes? [Y/n]:y
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/influx' to '/Users/emmanuelodeke/Desktop/exp/influx.pdf'
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/influx' to '/Users/emmanuelodeke/Desktop/exp/influx.txt'
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/few' to '/Users/emmanuelodeke/Desktop/exp/few.pdf'
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/few.docs' to '/Users/emmanuelodeke/Desktop/exp/few.docs.txt'
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/few.docs' to '/Users/emmanuelodeke/Desktop/exp/few.docs.odt'
Exported '/Users/emmanuelodeke/emm.odeke@gmail.com/test-exports/few.docs' to '/Users/emmanuelodeke/Desktop/exp/few.docs.pdf'
```

**Supported formats:**

* doc, docx
* jpeg, jpg
* gif
* html
* odt
* ods
* rtf
* pdf
* png
* ppt, pptx
* svg
* txt, text
* xls, xlsx

### Pushing

The `push` command uploads data to Google Drive to mirror data stored locally.

Like `pull`, you can run it without any arguments to push all of the files from the current path, or you can pass in one or more paths to push specific files or directories.

`push` also allows you to push content up to a desired traversal depth e.g

```shell
drive push -depth 1 head-folders
```

You can also push multiple paths that are children of the root of the mounted drive to a destination,

in relation to issue #612, using key `-destination`:

For example to push the content of `music/Travi$+Future`, `integrals/complex/compilations` directly to `a1/b2/c3`:

```shell
drive push -destination a1/b2/c3 music/Travi$+Future integrals/complex/compilations
```

To enable checksum verification during a push:

```shell
drive push -ignore-checksum=false
```

To keep your data encrypted at rest remotely on Google Drive:

```shell
drive push -encryption-password '$JiME5Umf' influx.txt
```
For E2E discussions, see [issue #543](https://github.com/odeke-em/issues/543):

drive also supports pushing content piped from stdin which can be accomplished by:

```shell
drive push -piped path
```

To selectively push by type e.g file vs directory/folder, you can use flags
- `files`
- `directories`

```shell
drive push -files a1/b2
drive push -directories tf1
```

Like most commands [.driveignore](#excluding-and-including-objects) can be used to filter which files to push.

+ Note: Use `drive push -hidden` to also push files starting with `.` like `.git`.

Here is an example using drive to backup the current working directory. It pushes a tar.gz archive created on the fly. No archive file is made on the machine running the command, so it doesn't waste disk space.

```shell
tar czf - . | drive push -piped backup-$(date +"%m-%d-%Y-"%T"").tar.gz
```

+ Note:
  * In response to [#107](https://github.com/odeke-em/drive/issues/107) and numerous other issues related to confusion about clashing paths, drive can now auto-rename clashing files. Use flag `-fix-clashes` during a `pull` or `push`, and drive will try to rename clashing files by adding a unique suffix at the end of the name, but right before the extension of a file (if the extension exists). If you haven't passed in the above `-fix-clashes` flag, drive will abort on trying to deal with clashing names. If you'd like to turn off this safety, pass in flag `-ignore-name-clashes`
  * In relation to [#57](https://github.com/odeke-em/drive/issues/57) and [@rakyll's #49](https://github.com/rakyll/drive/issues/49).
   A couple of scenarios in which data was getting totally clobbered and unrecoverable, drive now tries to play it safe and warn you if your data could potentially be lost e.g during a to-disk clobber for which you have no backup. At least with a push you have the luxury of untrashing content. To disable this safety, run drive with flag `-ignore-conflict` e.g:

    ```shell
    drive pull -ignore-conflict collaboration_documents
    ```

    Playing the safety card even more, if you want to get changes that are non clobberable ie only additions
    run drive with flag `-no-clobber` e.g:

    ```shell
    drive pull -no-clobber Makefile
    ```

  * Ordinarily your system will not traverse nested symlinks e.g:
  ```shell
    mkdir -p a/b
    mkdir -p ~/Desktop/z1/z2 && ls ~ > ~/Desktop/z1/z2/listing.txt
    ln -s ~/Desktop/z1/z2 a/b
    ls -R a # Should print only z2 and nothing inside it. 
  ```

    However in relation to [#80](https://github.com/odeke-em/drive/issues/80), for purposes of consistency with your Drive, traversing symlinks has been added.

For safety with non clobberable changes i.e only additions:

```shell
drive push -no-clobber
```

+ Due to the reasons above, drive should be able to warn you in case of total clobbers on data. To turn off this behaviour/safety, pass in the `-ignore-conflict` flag i.e:

```shell
drive push -force sure_of_content
```

To push without user input (i.e. without prompt)
```shell
drive push -quiet
```
or
```shell
drive push -no-prompt
```

To get Google Drive to convert a file to its native Google Docs format

```shell
drive push -convert
```
Extra features: to make Google Drive attempt Optical Character Recognition (OCR) for png, gif, pdf and jpg files.

```shell
drive push -ocr
```
Note: To use OCR, your account should have this feature. You can find out if your account has OCR allowed.

```shell
drive features
```

### Pulling And Pushing Notes

+ MimeType inference is from the file's extension.

  If you would like to coerce a certain mimeType that you'd prefer to assert with Google Drive pushes, use flag `-coerce-mime <short-key>` See [List of MIME type short keys](https://github.com/odeke-em/drive/wiki/List-of-MIME-type-short-keys) for the full list of short keys.

```shell
drive push -coerce-mime docx my_test_doc
```

+ Excluding certain operations can be done both for pull and push by passing in flag
`-exclude-ops` <csv_crud_values>

e.g

```shell
drive pull -exclude-ops "delete,update" vines
drive push -exclude-ops "create" sensitive_files
```

+ To show more information during pushes or pulls e.g show the current operation,
pass in option `-verbose` e.g:

```shell
drive pull -verbose 2015/Photos content
drive push -verbose Music Fall2014
```

+ In relation to issue #529, you can change the max retry counts for exponential backoff. Using a count < 0 falls back to the
default count of 20:
```shell
drive push -retry-count 4 a/bc/def terms
```

* You can also specify the upload chunk size to be used to push each file, by using flag
`-upload-chunk-size` whose value is in bytes. If you don't specify this flag, by default
the internal Google APIs use a value of 8MiB from constant `googleapi.DefaultUploadChunkSize`.
Please note that your value has to be a multiple of and atleast the minimum  upload chunksize
of 256KiB from constant `googleapi.MinUploadChunkSize`. See https://godoc.org/google.golang.org/api/googleapi#pkg-constants.
  If `-upload-chunk-size` is not set yet `-upload-rate-limit` is, `-upload-chunk-size` will be the same as `-upload-rate-limit`.

* To limit the upload bandwidth, please set `-upload-rate-limit=n`. It's in `n` KiB/s, default is unlimited.

### End to End Encryption

See [Issue #543](https://github.com/odeke-em/drive/issues/543)

This can be toggled when you supply a non-empty password ie

- `-encryption-password` for a push.
- `-decryption-password` for a pull.

When you supply argument `-encryption-password` during a push, drive will encrypt your data
and store it remotely encrypted(stored encrypted at rest), it can only be decrypted by you when you
perform a pull with the respective arg `-decryption-password`.

```shell
drive push -encryption-password '$400lsGO1Di3' few-ones.mp4 newest.mkv
```

```shell
drive pull -decryption-password '$400lsGO1Di3' few-ones.mp4 newest.mkv
```

If you supply the wrong password, you'll be warned if it cannot be decrypted

```shell
$ drive pull -decryption-password "4nG5troM" few-ones.mp4 newest.mkv
message corrupt or incorrect password
```

To pull normally push or pull your content, without attempting any *cryption attempts, skip
passing in a password and no attempts will be made.

### Publishing

The `pub` command publishes a file or directory globally so that anyone can view it on the web using the link returned.

```shell
drive pub photos
```

+ Publishing by fileId is also supported

```shell
drive pub -id 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU 0fM9rt0Yc9RTPSTZEanBsamZjUXM
```

### Unpublishing

The `unpub` command is the opposite of `pub`. It unpublishes a previously published file or directory.

```shell
drive unpub photos
```

+ Publishing by fileId is also supported

```shell
drive unpub -id 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU 0fM9rt0Yc9RTPSTZEanBsamZjUXM
```

### Sharing and Emailing

The `share` command enables you to share a set of files with specific users and assign them specific roles as well as specific generic access to the files. It also allows for email notifications on share.

```shell
drive share -emails odeke@ualberta.ca,odeke.ex@gmail.com -message "This is the substring file I told you about" -role reader,writer -type group mnt/substringfinder.c projects/kmp.c
$ drive share -emails emm.odeke@gmail.com,odeke@ualberta.ca -role reader,commenter -type user influx traversal/notes/conquest
```

For example to share a file with users of a mailing list and a custom message

```shell
drive share -emails drive-mailing-list@gmail.com -message "Here is the drive code" -role group mnt/drive
```

+ By default, an email notification is sent (even if -message is not specfified). To turn off email notification, use -notify=false

```shell
$ drive share -notify=false -emails emm.odeke@gmail.com,odeke@ualberta.ca -role reader,commenter -type user influx traversal/notes/conquest
```

+ The `share` command also supports sharing by fileId

```shell
drive share -emails developers@developers.devs -message "Developers, developers developers" -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

+ You can also share a file to only those with the link. As per [https://github.com/odeke-em/drive/issues/568](https://github.com/odeke-em/drive/issues/568), this file won't be publicly indexed. To turn this option on when sharing the file,
use flag `-with-link`.

```shell
drive share -with-link ComedyPunchlineDrumSound.mp3
```

### Unsharing

The `unshare` command revokes access of a specific accountType to a set of files.

When no -role is given it by default assumes you want to revoke all access ie "reader", "writer", "commenter"

```shell
drive unshare -type group mnt/drive
drive unshare -emails  emm.odeke@gmail.com,odeke@ualberta.ca -type user,group -role reader,commenter infinity newfiles/confidential
```

+ Also supports unsharing by fileId

```shell
drive unshare -type group -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Starring Or Unstarring

To star or unstar documents,

```shell
drive star information quest/A/B/C
drive star -id 0fM9rt0Yc9RTPaDdsNzg1dXVjM0E 0fM9rt0Yc9RTPaTVGc1pzODN1NjQ 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU
```

```shell
drive unstar information quest/A/B/C
drive unstar -id 0fM9rt0Yc9RTPaDdsNzg1dXVjM0E 0fM9rt0Yc9RTPaTVGc1pzODN1NjQ 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU
```

### Diffing

The `diff` command compares local files with their remote equivalents. It allows for multiple paths to be passed in e.g

```shell
drive diff changeLogs.log notes sub-folders/
```

You can diff to a desired depth

```shell
drive diff -depth 2 sub-folders/ contacts/ listings.txt
```

You can also switch the base, either local or remote by using flag `-base-local`

```shell
drive diff -base-local=true assignments photos # To use local as the base
drive diff -base-local=false infocom photos # To use remote as the base
```

You can only diff for short changes that is only name differences, file modTimes and types, you can use flag `-skip-content-check`.

```shell
drive diff -skip-content-check
```

### Touching

Files that exist remotely can be touched i.e their modification time updated to that on the remote server using the `touch` command:

```shell
drive touch Photos/img001.png logs/log9907.txt
```

For example to touch all files that begin with digits 0  to 9:

```shell
drive touch -matches $(seq 0 9)
```

+ Also supports touching of files by fileId

```shell
drive touch -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

+ You can also touch files to a desired depth of nesting within their parent folders.

```shell
drive touch -depth 3 mnt newest flux
drive touch -depth -1 -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
drive touch -depth 1 -matches $(seq 0 9)
```

+ You can also touch and explicitly set the modification time for files by:
```shell
drive touch -time 20120202120000 ComedyPunchlineDrumSound.mp3
/share-testing/ComedyPunchlineDrumSound.mp3: 2012-02-02 12:00:00 +0000 UTC
```

+ Specify the time format that you'd like to use when specifying the time e.g
```shell
drive touch -format "2006-01-02-15:04:05.0000Z" -time "2016-02-03-08:12:15.0070Z" outf.go
/share-testing/outf.go: 2016-02-03 08:12:15 +0000 UTC
```
The mentioned time format has to be relative to how you would represent
"Mon Jan 2 15:04:05 -0700 MST 2006".
See the documentation for time formatting here [time.Parse](https://golang.org/pkg/time/#Parse)

+ Specify the touch time offset from the clock on your machine where:
- minus(-) means ago e.g 30 hours ago -> -30h
- blank or plus(+) means from now e.g 10 minutes -> 10m or +10m
```shell
drive touch -duration -30h ComedyPunchlineDrumSound.mp3 outf.go
/share-testing/outf.go: 2016-09-10 08:06:39 +0000 UTC
/share-testing/ComedyPunchlineDrumSound.mp3: 2016-09-10 08:06:39 +0000 UTC
```

### Trashing And Untrashing

Files can be trashed using the `trash` command:

```shell
drive trash Demo
```

To trash files that contain a prefix match e.g all files that begin with Untitled, or Make

Note: This option uses the current working directory as the parent that the paths belong to.

```shell
drive trash -matches Untitled Make
```

Files that have been trashed can be restored using the `untrash` command:

```shell
drive untrash Demo
```

To untrash files that match a certain prefix pattern

```shell
drive untrash -matches pQueue photos Untitled
```

+ Also supports trashing/untrashing by fileId

```shell
drive trash -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
drive untrash -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Emptying The Trash

Emptying the trash will permanently delete all trashed files. Caution: They cannot be recovered after running this command.

```shell
drive emptytrash
```

### Deleting

Deleting items will PERMANENTLY remove the items from your drive. This operation is irreversible.

```shell
drive delete flux.mp4
```

```shell
drive delete -matches onyx swp
```

+ Also supports deletion by fileIds

```shell
drive delete -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Listing

The `list` command shows a paginated list of files present remotely.

Run it without arguments to list all files in the current directory's remote equivalent:

```shell
drive list
```

Pass in a directory path to list files in that directory:

```shell
drive list photos
```

To list matches

```shell
drive list -matches mp4 go
```

The `-trashed` option can be specified to show trashed files in the listing:

```shell
drive list -trashed photos
```

To get detailed information about the listings e.g owner information and the version number of all listed files:

```shell
drive list -owners -l -version
```

+ Also supports listing by fileIds

```shell
drive list -depth 3 -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

+ Listing allows for sorting by fields e.g `name`, `version`, `size, `modtime`, lastModifiedByMeTime `lvt`, `md5`. To do this in reverse order, suffix `_r` or `-` to the selected key

e.g to first sort by modTime, then largest-to-smallest and finally most number of saves:

```
drive list -sort modtime,size_r,version_r Photos
```

* For advanced listing

```shell
drive list -skip-mime mp4,doc,txt
drive list -match-mime xls,docx
drive list -exact-title url_test,Photos
```

### Stating

The `stat` commands show detailed file information for example people with whom it is shared, their roles and accountTypes, and
fileId etc. It is useful to help determine whom and what you want to be set when performing share/unshare

```shell
drive stat mnt
```

By default `stat` won't recursively stat a directory, to enable recursive stating:

```shell
drive stat -r mnt
```

+ Also supports stat-ing by fileIds

```shell
drive stat -r -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

OR

```shell
drive stat -depth 4 -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Printing URL

The url command prints out the url of a file. It allows you to specify multiple paths relative to root or even by id

```shell
drive url Photos/2015/07/Releases intros/flux
drive url -id  0Bz5qQkvRAeVEV0JtZl4zVUZFWWx  1Pwu8lzYc9RTPTEpwYjhRMnlSbDQ 0Cz5qUrvDBeX4RUFFbFZ5UXhKZm8
```

### Editing Description

You can edit the description of a file like this

```shell
drive edit-desc -description "This is a new file description" freshFolders/1.txt commonCore/
drive edit-description -description "This is a new file description" freshFolders/1.txt commonCore/
```

Even more conveniently by piping content

```shell
cat fileDescriptions | drive edit-desc -piped  targetFile influx/1.txt
```

### Retrieving MD5 Checksums

The `md5sum` command quickly retrieves the md5 checksums of the files on your drive. The result can be fed into the "md5sum -c" shell command to validate the integrity of the files on Drive versus the local copies.

Check that files on Drive are present and match local files:

```shell
~/MyDrive/folder$ drive md5sum | md5sum -c
```

Do a two-way diff (will also locate files missing on either side)

```shell
~/MyDrive/folder$ diff <(drive md5sum) <(md5sum *)
```

Same as above, but include subfolders 

```shell
~/MyDrive/folder$ diff <(drive md5sum -r) <(find * -type f | sort | xargs md5sum)
```

Compare across two different Drive accounts, including subfolders

```shell
~$ diff <(drive md5sum -r MyDrive/folder) <(drive md5sum -r OtherDrive/otherfolder)
```

* Note: Running the 'drive md5sum' command retrieves pre-computed md5 sums from Drive; its speed is proportional to the number of files on Drive. Running the shell 'md5sum' command on local files requires reading through the files; its speed is proportional to the size of the files._

### Retrieving FileId

You can retrieve just the fileId for specified paths
```shell
drive id [-depth n] [paths...]
drive file-id [-depth n] [paths...]
```

For example:

```shell
drive file-id -depth 2 dup-tests bug-reproductions
# drive file-id -depth 2 dup-tests bug-reproductions
FileId                                           Relative Path
"0By5qKlgRJeV2NB1OTlpmSkg8TFU"                   "/dup-tests"
"0Bz5wQlgRJeP2QkRSenBTaUowU3c"                   "/dup-tests/influx_0"
"0Cu5wQlgRJeV2d2VmY29HV217TFE"                   "/dup-tests/a"
"0Cy5wQlgRJeX2WXVFMnQyQ2NDRTQ"                   "/dup-tests/influx"
"0Cy5wQlgRJeP2YGMiOC15OEpUZnM"                   "/bug-reproductions"
"0Cy5wQlgRJeV2MzFtTm50NVV5NW8"                   "/bug-reproductions/drive-406"
"1xmXPziMPEgq2dK-JqaUytKz_By8S_7_RVY79ceRoZwv"	 "info-bulletins"
```

### Retrieving Quota

The `quota` command prints information about your drive, such as the account type, bytes used/free, and the total amount of storage available.

```shell
drive quota
```

### Retrieving Features

The `features` command provides information about the features present on the
drive being queried and the request limit in queries per second

```shell
drive features
```

### Creating

drive allows you to create an empty file or folder remotely
Sample usage:

```shell
drive new -folder flux
drive new -mime-key doc bofx
drive new -mime-key folder content
drive new -mime-key presentation ProjectsPresentation
drive new -mime-key sheet Hours2015Sept
drive new -mime-key form taxForm2016 taxFormCounty
drive new flux.txt oxen.pdf # Allow auto type resolution from the extension
```

### Opening

The open command allows for files to be opened by the default file browser, default web browser, either by path or by id for paths that exist atleast remotely

```shell
drive open -file-browser=false -web-browser f1/f2/f3 jamaican.mp4
drive open -file-browser -id 0Bz8qQkpZAeV9T1PObvs2Y3BMQEj 0Y9jtQkpXAeV9M1PObvs4Y3BNRFk
```

### Copying

drive allows you to copy content remotely without having to explicitly download and then reupload.

```shell
drive copy -r blobStore.py mnt flagging
```

```shell
drive copy blobStore.py blobStoreDuplicated.py
```

+ Also supports copying by fileIds

```shell
drive copy -r -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U ../content
```

### Moving

drive allows you to move content remotely between folders. To do so:

```shell
drive move photos/2015 angles library archives/storage
```

+ Also supports moving by fileId

```shell
drive move -id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U ../../new_location
```

Google Drive supports multi-parent folder structure, where one file/folder can be placed in more than one parent folder.
It consumes no extra disk space on the Cloud, but after pulling such structure it may double your files several times in your file structure.
Pushing non deduplicated folder structures back may also break things, so be careful.

To place file/folder into new parent folder, keeping old one as well, use `-keep-parent` option

```shell
$ drive move -keep-parent photos/2015 angles library second_parent_folder
```

### Renaming

drive allows you to rename a file/folder remotely.
Two arguments are required to rename ie `<relativePath/To/source or Id>` `<newName>`.

To perform a rename:

```shell
drive rename url_test url_test_results
drive rename openSrc/2015 2015-Contributions
```

+ Also supports renaming by fileId

```shell
drive rename 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 fluxing
```

To turn off renaming locally or remotely, use flags
`-local=false` or `-remote=false`. By default both are turned on.

For example

```shell
drive rename -local=false -remote=true a/b/c/d/e/f flux
```

### Command Aliases

`drive` supports a few aliases to make usage familiar to the utilities in your shell e.g:
+ cp : copy
+ ls : list 
+ mv : move
+ rm : delete

### Detecting And Fixing Clashes

You can deal with clashes by using command `drive clashes`.

* To list clashes, you can do

```shell
drive clashes [-depth n] [paths...]
drive clashes -list [-depth n] [paths...] # To be more explicit
```

* To fix clashes, you can do:

```
drive clashes -fix [-fix-mode mode] [-depth n] [paths...]
```

There are two available modes for `-fix-mode`:
  * `rename`: this is the default behavior
  * `trash`: trashing *both* new and old files

## .desktop Files

As previously mentioned, Google Docs, Drawings, Presentations, Sheets etc and all files affiliated
with docs.google.com cannot be downloaded raw but only exported. Due to popular demand, Linux users
desire the ability to have \*.desktop files that enable the file to be opened appropriately by an external opener.
Thus by default on Linux, drive will create \*.desktop files for files that fall into this category.

To turn off this behavior, you can set flag `-desktop-links` to false e.g
```shell
drive pull -desktop-links=false
```

### Fetching And Pruning Missing Index Files

* index 

If you would like to fetch missing index files for files that would otherwise not need any modifications, run:

```shell
drive index path1 path2 path3/path3.1 # To fetch any missing indices in those paths
drive index -id 0CLu4lbUI9RTRM80k8EMoe5JQY2z
```

You can also fetch specific files by prefix matches
```shell
drive index -matches mp3 jpg
```

* prune

In case you might have deleted files remotely but never using drive, and feel like you have stale indices,
running `drive index -prune` will search your entire indices dir for index files that do not exist remotely and remove those ones

```shell
drive index -prune
```

* prune-and-index
To combine both operations (prune and then fetch) for indices:

```shell
drive index -all-ops
```

### Drive server

To enable services like qr-code sharing, you'll need to have the server running that will serve content once invoked in a web browser to allow for resources to be accessed on another device e.g your mobile phone

```shell
go get github.com/odeke-em/drive/drive-server && drive-server
drive-server
```

Pre-requisites:
  + DRIVE\_SERVER\_PUB\_KEY
  + DRIVE\_SERVER\_PRIV\_KEY

Optionally
  + DRIVE\_SERVER\_PORT : default is 8010
  + DRIVE\_SERVER\_HOST : default is localhost

If the above keys are not set in your env, you can do this

```shell
DRIVE_SERVER_PUB_KEY=<pub_key> DRIVE_SERVER_PRIV_KEY=<priv_key> [DRIVE...] drive-server
```

### QR Code Share

Instead of traditionally copying long links, drive can now allow you to share a link to a file by means of a QR code that is generated after a redirect through your web browser. 

From then on, you can use your mobile device or any other QR code reader to get to that file.
In order for this to run, you have to have the `drive-server` running

As long as the server is running on a known domain, then you can start the qr-link getting ie

```shell
drive qr vines/kevin-hart.mp4 notes/caches.pdf
drive qr -address http://192.168.1.113:8010 books/newest.pdf maps/infoGraphic.png
drive qr -address https://my.server books/newest.pdf maps/infoGraphic.png
```

That should open up a browser with the QR code that when scanned will open up the desired file.

### About

The `about` command provides information about the program as well as that about
your Google Drive. Think of it as a hybrid between the `features` and `quota` commands.
```shell
drive about
```

OR for detailed information
```shell
drive about -features -quota
```

### Help

Run the `help` command without any arguments to see information about the commands that are available:

```shell
drive help
```

Pass in the name of a command to get information about that specific command and the options that can be passed to it.

```shell
drive help push
```

To get help for all the commands
```shell
drive help all
```

### Filing Issues

In case of any issue, you can file one by using command `issue` aka `report-issue` aka `report`.
It takes flags `-title` `-body` `-piped`.

* If `-piped` is set, it expects to read the body from standard input.

A successful issue-filing request will open up the project's issue tracker in your web browser.

```
drive issue -title "Can't open my file" -body "Drive trips out every time"
drive report-issue -title "Can't open my file" -body "Drive trips out every time"
cat bugReport.txt | drive issue -piped -title "push: dump on pushing from this directory"
```

### Revoking Account Access

To revoke OAuth Access of drive to your account, when logged in with your Google account, go to https://security.google.com/settings/security/permissions and revoke the desired permissions

### Uninstalling

To remove `drive` from your computer, you'll need to take out:
+ $GOPATH/bin/drive
+ $GOPATH/src/github.com/odeke-em/drive
+ $GOPATH/pkg/github.com/odeke-em/drive
+ $GOPATH/pkg/github.com/odeke-em/drive.a

* Also do not forget to revoke drive's access in case you need to uninstall it.

## Applying Patches 
To  apply patches of code e.g in the midst of bug fixes, you'll just need a little bit of git fiddling.

For example to patch your code with that on remote branch patch-1, you'll need to go into the source
code directory, fetch all content from the git remote, checkout the patch branch then run the go installation: something like this.

```shell
cd $GOPATH/src/github.com/odeke-em/drive
git fetch --all
git checkout patch-1
git pull origin patch-1
go get github.com/odeke-em/drive/cmd/drive
```

## Why Another Google Drive Client?

Background sync is not just hard, it is stupid. Here are my technical and philosophical rants about why it is not worth to implement:

* Too racy. Data is shared between your remote resource, local disk and sometimes in your sync daemon's in-memory structs. Any party could touch a file at any time. It is hard to lock these actions. You end up working with multiple isolated copies of the same file and trying to determine which is the latest version that should be synced across different contexts.

* It requires great scheduling to perform best with your existing environmental constraints. On the other hand, file attribute have an impact on the sync strategy. Large files block -- you wouldn't like to sit on and wait for a VM image to get synced before you can start working on a tiny text file.

* It needs to read your mind to understand your priorities. Which file do you need most? It needs to read your mind to foresee your future actions. I'm editing a file, and saving the changes time to time. Why not to wait until I feel confident enough to commit the changes remotely?

`drive` is not a sync daemon, it provides:

* Upstreaming and downstreaming. Unlike a sync command, we provide pull and push actions. The user has the opportunity to decide what to do with their local copy and when they decide to. Make some changes, either push the file remotely or revert it to the remote version. You can perform these actions with user prompt:

	    echo "hello" > hello.txt
	    drive push # pushes hello.txt to Google Drive
	    echo "more text" >> hello.txt
	    drive pull # overwrites the local changes with the remote version

* Allowing to work with a specific file or directory, optionally not recursively. If you recently uploaded a large VM image to Google Drive, yet only a few text files are required for you to work, simply only push/pull the exact files you'd like to worth with:

	    echo "hello" > hello.txt
	    drive push hello.txt # pushes only the specified file
	    drive pull path/to/a/b path2/to/c/d/e # pulls the remote directory recursively

* Better I/O scheduling. One of the major goals is to provide better scheduling to improve upload/download times.

* Possibility to support multiple accounts. Pull from or push to multiple Google Drive remotes. Possibility to support multiple backends. Why not to push to Dropbox or Box as well?

## Known Issues

* It probably doesn't work on Windows.
* Google Drive allows a directory to contain files/directories with the same name. Client doesn't handle these cases yet. We don't recommend you to use `drive` if you have such files/directories to avoid data loss.
* Racing conditions occur if remote is being modified while we're trying to update the file. Google Drive provides resource versioning with ETags, use Etags to avoid racy cases.
* drive rejects reading from namedPipes because they could infinitely hang. See [issue #208](https://github.com/odeke-em/drive/issues/208).

## Reaching Out

Doing anything interesting with drive or want to share your favorite tips and tricks? Check out the [wiki](https://github.com/odeke-em/drive/wiki) and feel free to reach out with ideas for features or requests.

## Disclaimer

This project is not supported nor maintained by Google.

## LICENSE

Copyright 2013 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
