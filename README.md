# drive

[![Build Status](https://travis-ci.org/odeke-em/drive.png?branch=master)](https://travis-ci.org/odeke-em/drive)

`drive` is a tiny program to pull or push [Google Drive](https://drive.google.com) files.

`drive` was originally developed by [Burcu Dogan](https://github.com/rakyll) while working on the Google Drive team. This repository contains the latest version of the code, as she is no longer able to maintain it.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
  - [Platform Packages](#platform-packages)
  - [Godep](#godep)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Initializing](#initializing)
  - [De Initializing](#de-initializing)
  - [Traversal Depth](#traversal-depth)
  - [Pulling](#pulling)
    - [Exporting Docs](#exporting-docs)
  - [Pushing](#pushing)
  - [Publishing](#publishing)
  - [Unpublishing](#unpublishing)
  - [Sharing and Emailing](#sharing-and-emailing)
  - [Unsharing](#unsharing)
  - [Diffing](#diffing)
  - [Touching](#touching)
  - [Trashing and Untrashing](#trashing-and-untrashing)
  - [Emptying the Trash](#emptying-the-trash)
  - [Deleting](#deleting)
  - [Listing Files](#listing-files)
  - [Stating Files](#stating-files)
  - [Retrieving md5 checksums](#retrieving-md5-checksums)
  - [Editing Description](#editing-description)
  - [New File](#new-file)
  - [Quota](#quota)
  - [Features](#features)
  - [About](#about)
  - [Help](#help)
  - [Move](#move)
  - [Rename](#rename)
  - [DriveIgnore](#driveignore)
  - [DesktopEntry](#desktopentry)
  - [Command Aliases](#command-aliases)
  - [Index Prune](#index-prune)
  - [Url](#url)
- [Revoking Account Access](#revoking-account-access)
- [Uninstalling](#uninstalling)
- [Applying patches](#applying-patches)
- [Why another Google Drive client?](#why-another-google-drive-client)
- [Known issues](#known-issues)
- [Reach out](#reach-out)
- [Disclaimer](#disclaimer)
- [LICENSE](#license)

## Requirements

go 1.3.X or higher is required. See [here](https://golang.org/doc/install) for installation instructions and platform installers.

* Make sure to set your GOPATH in your env, .bashrc or .bash\_profile file. If you have not yet set it, you can do so like this:

```shell
$ cat << ! >> ~/.bashrc
> export GOPATH=\$HOME/gopath
> export PATH=\$GOPATH:\$GOPATH/bin:\$PATH
> !
$ source ~/.bashrc # To reload the settings and get the newly set ones # Or open a fresh terminal
```
The above setup will ensure that the drive binary after compilation can be invoked from your current path.

## Installation

To install from the latest source, run:

```shell
$ go get -u github.com/odeke-em/drive/cmd/drive
```

Otherwise:

* In order to address [issue #138](https://github.com/odeke-em/drive/issues/138), where debug information should be bundled with the binary, you'll need to run:

```shell
$ go get github.com/odeke-em/drive/drive-gen && drive-gen
```

In case you need a specific binary e.g for Debian folks [issue #271](https://github.com/odeke-em/drive/issues/271) and [issue 277](https://github.com/odeke-em/drive/issues/277)

```shell
$ go get -u github.com/odeke-em/drive/drive-google
```

That should produce a binary `drive-google`

OR

To bundle debug information with the binary, you can run:

```shell
$ go get -u github.com/odeke-em/drive/drive-gen && drive-gen drive-google
```


### Godep

+ Using godep
```
$ cd $GOPATH/src/github.com/odeke-em/drive/drive-gen && godep save
```

+ Unravelling/Restoring dependencies
```
$ cd $GOPATH/src/github.com/odeke-em/drive/drive-gen && godep restore
```

Please see file `drive-gen/README.md` for more information.


### Platform Packages

For curated packages on your favorite platform, please see file [Platform Packages.md](https://github.com/odeke-em/drive/blob/master/platform_packages.md).

Is your platform missing a package? Feel free to prepare / contribute an installation package and then submit a PR to add it in.

## Configuration

Optionally set the `GOOGLE_API_CLIENT_ID` and `GOOGLE_API_CLIENT_SECRET` environment variables to use your own API keys.

## Usage

### Initializing

Before you can use `drive`, you'll need to mount your Google Drive directory on your local file system:

```shell
$ drive init ~/gdrive
$ cd ~/gdrive
```

### De Initializing

The opposite of `drive init`, it will remove your credentials locally as well as configuration associated files.

```shell
$ drive deinit [--no-prompt]
```

For a complete deinit-ialization, don't forget to revoke account access, [please see revoking account access](#revoking-account-access)


### Traversal Depth

Before talking about the features of drive, it is useful to know about "Traversal Depth".

Throughout this README the usage of the term "Traversal Depth" refers to the number of

nodes/hops/items that it takes to get from one parent to children. In the options that allow it, you'll have a flag option `--depth <n>` where n is an integer

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



### Pulling

The `pull` command downloads data that does not exist locally but does remotely on Google drive, and may delete local data that is not present on Google Drive. 
Run it without any arguments to pull all of the files from the current path:

```shell
$ drive pull
```

Pulling by matches is also supported

```shell
$ cd ~/myDrive/content/2015
$ drive pull --matches vines docx
```

To force download from paths that otherwise would be marked with no-changes

```shell
$ drive pull -force
```

To pull specific files or directories, pass in one or more paths:

```shell
$ drive pull photos/img001.png docs
```

Pulling by id is also supported

```shell
$ drive pull --id 0fM9rt0Yc9RTPaDdsNzg1dXVjM0E 0fM9rt0Yc9RTPaTVGc1pzODN1NjQ 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU
```

`pull` optionally allows you to pull content up to a desired depth.

Say you would like to get just folder items until the second level

```shell
$ drive pull --depth 2 heavy-files summaries
```

Traverse deep to infinity and beyond

```shell
$ drive pull --depth -1 all-my-files
```


## Note: Checksum verification:

* By default checksum-ing is turned off because it was deemed to be quite vigorous and unnecessary for most cases.
* Due to popular demand to cover the common case in which size + modTime differences are sufficient to detect file changes. The discussion stemmed from issue [#117](https://github.com/odeke-em/drive/issues/117).

   To turn checksum verification back on:

```shell
$ drive pull -ignore-checksum=false
```


drive also supports piping pulled content to stdout which can be accomplished by:

```shell
$ drive pull -piped path1 path2
```

#### Exporting Docs

By default, the `pull` command will export Google Docs documents as PDF files. To specify other formats, use the `-export` option:

```shell
$ drive pull -export pdf,rtf,docx,txt
```

To explicitly export instead of using `--force`

```shell
$ drive pull --export pdf,rtf,docx,txt --explicitly-export
```

By default, the exported files will be placed in a new directory suffixed by `\_exports` in the same path. To export the files to a different directory, use the `-export-dir` option:

```shell
$ drive pull -export pdf,rtf,docx,txt -export-dir ~/Desktop/exports
```

**Supported formats:**

* doc, docx
* jpeg, jpg
* gif
* html
* odt
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
$ drive push --depth 1 head-folders
```

Note: To ignore checksum verification during a push:

```shell
$ drive push -ignore-checksum
```

drive also supports pushing content piped from stdin which can be accomplished by:

```shell
$ drive push -piped path
```

Here is an example using drive to backup the current working directory. It pushes a tar.gz archive created on the fly. No archive file is made on the machine running the command, so it doesn't waste disk space.

```shell
$ tar czf - . | drive push -piped backup-$(date +"%m-%d-%Y-"%T"").tar.gz
```

+ Note:
  * In response to [#107](https://github.com/odeke-em/drive/issues/107) and numerous other issues related to confusion about clashing paths, drive can now auto-rename clashing files. Use flag `--fix-clashes` during a `pull` or `push`, and drive will try to rename clashing files by adding a unique suffix at the end of the name, but right before the extension of a file (if the extension exists). If you haven't passed in the above `--fix-clashes` flag, drive will abort on trying to deal with clashing names. If you'd like to turn off this safety, pass in flag `--ignore-name-clashes`
  * In relation to [#57](https://github.com/odeke-em/drive/issues/57) and [@rakyll's #49](https://github.com/rakyll/drive/issues/49).
   A couple of scenarios in which data was getting totally clobbered and unrecoverable, drive now tries to play it safe and warn you if your data could potentially be lost e.g during a to-disk clobber for which you have no backup. At least with a push you have the luxury of untrashing content. To disable this safety, run drive with flag `-ignore-conflict` e.g:

    ```shell
    $ drive pull -ignore-conflict collaboration_documents
    ```

    Playing the safety card even more, if you want to get changes that are non clobberable ie only additions
    run drive with flag `-no-clobber` e.g:

    ```shell
    $ drive pull -no-clobber Makefile
    ```

  * Ordinarily your system will not traverse nested symlinks e.g:
  ```shell
    $ mkdir -p a/b
    $ mkdir -p ~/Desktop/z1/z2 && ls ~ > ~/Desktop/z1/z2/listing.txt
    $ ln -s ~/Desktop/z1/z2 a/b
    $ ls -R a # Should print only z2 and nothing inside it. 
  ```

    However in relation to [#80](https://github.com/odeke-em/drive/issues/80), for purposes of consistency with your Drive, traversing symlinks has been added.

For safety with non clobberable changes i.e only additions:

```shell
$ drive push -no-clobber
```

+ Due to the reasons above, drive should be able to warn you in case of total clobbers on data. To turn off this behaviour/safety, pass in the `-ignore-conflict` flag i.e:

```shell
$ drive push -force sure_of_content
```

To pull without user input (i.e. without prompt)
```shell
$ drive push -quiet
```
or
```shell
$ drive push -no-prompt
```

To get Google Drive to convert a file to its native Google Docs format

```shell
$ drive push -convert
```
Extra features: to make Google Drive attempt Optical Character Recognition (OCR) for png, gif, pdf and jpg files.

```shell
$ drive push -ocr
```
Note: To use OCR, your account should have this feature. You can find out if your account has OCR allowed.

```shell
$ drive features
```

## Note:

+ MimeType inference is from the file's extension.

  If you would like to coerce a certain mimeType that you'd prefer to assert with Google Drive pushes, use flag `-coerce-mime <short-key>` See [List of MIME type short keys](https://github.com/odeke-em/drive/wiki/List-of-MIME-type-short-keys) for the full list of short keys.

```shell
$ drive push -coerce-mime docx my_test_doc
```

+ Excluding certain operations can be done both for pull and push by passing in flag
`--exclude-ops` <csv_crud_values>

e.g

```shell
$ drive pull --exclude-ops "delete,update" vines
$ drive push --exclude-ops "create" sensitive_files
```

+ To show more information during pushes or pulls e.g show the current operation,
pass in option `--verbose` e.g:

```shell
$ drive pull --verbose 2015/Photos content
$ drive push --verbose Music Fall2014
```

### Publishing

The `pub` command publishes a file or directory globally so that anyone can view it on the web using the link returned.

```shell
$ drive pub photos
```

+ Publishing by fileId is also supported

```shell
$ drive pub --id 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU 0fM9rt0Yc9RTPSTZEanBsamZjUXM
```

### Unpublishing

The `unpub` command is the opposite of `pub`. It unpublishes a previously published file or directory.

```shell
$ drive unpub photos
```

+ Publishing by fileId is also supported

```shell
$ drive unpub --id 0fM9rt0Yc9RTPV1NaNFp5WlV3dlU 0fM9rt0Yc9RTPSTZEanBsamZjUXM
```

### Sharing and Emailing

The `share` command enables you to share a set of files with specific users and assign them specific roles as well as specific generic access to the files. It also allows for email notifications on share.

```shell
$ drive share -emails odeke@ualberta.ca,odeke.ex@gmail.com -message "This is the substring file I told you about" -role writer -type group mnt/substringfinder.c projects/kmp.c
```

For example to share a file with users of a mailing list and a custom message

```shell
$ drive share -emails drive-mailing-list@gmail.com -message "Here is the drive code" -role group mnt/drive
```

+ Also supports sharing by fileId

```shell
$ drive share --emails developers@developers.devs --message "Developers, developers developers" --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Unsharing

The `unshare` command revokes access of a specific accountType to a set of files.

```shell
$ drive unshare -type group mnt/drive
```

+ Also supports unsharing by fileId

```shell
$ drive unshare --type group --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Diffing

The `diff` command compares local files with their remote equivalents. It allows for multiple paths to be passed in e.g

```shell
$ drive diff changeLogs.log notes sub-folders/
```

You can diff to a desired depth

```shell
$ drive diff --depth 2 sub-folders/ contacts/ listings.txt
```


### Touching

Files that exist remotely can be touched i.e their modification time updated to that on the remote server using the `touch` command:

```shell
$ drive touch Photos/img001.png logs/log9907.txt
```

For example to touch all files that begin with digits 0  to 9:

```shell
$ drive touch -matches $(seq 0 9)
```

+ Also supports touching of files by fileId

```shell
$ drive touch --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Trashing and Untrashing

Files can be trashed using the `trash` command:

```shell
$ drive trash Demo
```

To trash files that contain a prefix match e.g all files that begin with Untitled, or Make

Note: This option uses the current working directory as the parent that the paths belong to.

```shell
$ drive trash -matches Untitled Make
```

Files that have been trashed can be restored using the `untrash` command:

```shell
$ drive untrash Demo
```

To untrash files that match a certain prefix pattern

```shell
$ drive untrash -matches pQueue photos Untitled
```

+ Also supports trashing/untrashing by fileId

```shell
$ drive trash --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
$ drive untrash --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```


### Emptying the Trash

Emptying the trash will permanently delete all trashed files. Caution: They cannot be recovered after running this command.

```shell
$ drive emptytrash
```

### Deleting

Deleting items will PERMANENTLY remove the items from your drive. This operation is irreversible.

```shell
$ drive delete flux.mp4
```

```shell
$ drive delete --matches onyx swp
```

+ Also supports deletion by fileIds

```shell
$ drive delete --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```


### Listing Files

The `list` command shows a paginated list of files present remotely.

Run it without arguments to list all files in the current directory's remote equivalent:

```shell
$ drive list
```

Pass in a directory path to list files in that directory:

```shell
$ drive list photos
```

To list matches

```shell
$ drive list --matches mp4 go
```

The `-trashed` option can be specified to show trashed files in the listing:

```shell
$ drive list -trashed photos
```

To get detailed information about the listings e.g owner information and the version number of all listed files:

```shell
$ drive list -owners -l -version
```

+ Also supports listing by fileIds

```shell
$ drive list -depth 3 --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

+ Listing allows for sorting by fields e.g `name`, `version`, `size, `modtime`, lastModifiedByMeTime `lvt`, `md5`. To do this in reverse order, suffix `_r` or `-` to the selected key

e.g to first sort by modTime, then largest-to-smallest and finally most number of saves:

```
$ drive list --sort modtime,size_r,version_r Photos
```

* For advanced listing

```shell
$ drive list --skip-mime mp4,doc,txt
$ drive list --match-mime xls,docx
$ drive list --exact-title url_test,Photos
```

### Stating Files

The `stat` commands show detailed file information for example people with whom it is shared, their roles and accountTypes, and
fileId etc. It is useful to help determine whom and what you want to be set when performing share/unshare

```shell
$ drive stat mnt
```

By default `stat` won't recursively stat a directory, to enable recursive stating:

```shell
$ drive stat -r mnt
```

+ Also supports stat-ing by fileIds

```shell
$ drive stat -r --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

OR

```shell
$ drive stat -depth 4 --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U
```

### Retrieving md5 Checksums

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


### New File

drive allows you to create an empty file or folder remotely
Sample usage:

```shell
$ drive new --folder flux
$ drive new --mime-key doc bofx
$ drive new --mime-key folder content
$ drive new --mime-key presentation ProjectsPresentation
$ drive new --mime-key sheet Hours2015Sept
$ drive new --mime-key form taxForm2016 taxFormCounty
$ drive new flux.txt oxen.pdf # Allow auto type resolution from the extension
```

### Editing Description

You can edit the description of a file like this

```shell
$ drive edit-desc --description "This is a new file description" freshFolders/1.txt commonCore/
$ drive edit-description --description "This is a new file description" freshFolders/1.txt commonCore/
```

Even more conveniently by piping content

```shell
$ cat fileDescriptions | drive edit-desc --piped  targetFile influx/1.txt
```

### Quota

The `quota` command prints information about your drive, such as the account type, bytes used/free, and the total amount of storage available.

```shell
$ drive quota
```

### Features

The `features` command provides information about the features present on the
drive being queried and the request limit in queries per second

```shell
$ drive features
```

### About

The `about` command provides information about the program as well as that about
your Google Drive. Think of it as a hybrid between the `features` and `quota` commands.
```shell
$ drive about
```

OR for detailed information
```shell
$ drive about -features -quota
```

### Help

Run the `help` command without any arguments to see information about the commands that are available:

```shell
$ drive help
```

Pass in the name of a command to get information about that specific command and the options that can be passed to it.

```shell
$ drive help push
```

To get help for all the commands
```shell
$ drive help all
```

### Copying

drive allows you to copy content remotely without having to explicitly download and then reupload.

```shell
$ drive copy -r blobStore.py mnt flagging
```

```shell
$ drive copy blobStore.py blobStoreDuplicated.py
```

+ Also supports copying by fileIds

```shell
$ drive copy -r --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U ../content
```


### Rename

drive allows you to rename a file/folder remotely. To do so:

```shell
$ drive rename url_test url_test_results
$ drive rename openSrc/2015 2015-Contributions
```

+ Also supports renaming by fileId

```shell
$ drive rename 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 fluxing
```


### Move

drive allows you to move content remotely between folders. To do so:

```shell
$ drive move photos/2015 angles library archives/storage
```

+ Also supports moving by fileId

```shell
$ drive move --id 0fM9rt0Yc9RTPeHRfRHRRU0dIY97 0fM9rt0Yc9kJRPSTFNk9kSTVvb0U ../../new_location
```


### DriveIgnore

drive allows you to specify a '.driveignore' file similar to your .gitignore, in the root
directory of the mounted drive. Blank lines and those prefixed by '#' are considered as comments and skipped.

For example:

```shell
$ cat << $ >> .driveignore
> # My drive ignore file
> \.gd$
> \.so$
> \.swp$
> $
```

Note: Pattern matching and suffixes are done by regular expression matching so make sure to use a valid regular expression suffix.

## DesktopEntry

As previously mentioned, Google Docs, Drawings, Presentations, Sheets etc and all files affiliated
with docs.google.com cannot be downloaded raw but only exported. Due to popular demand, Linux users
desire the ability to have \*.desktop files that enable the file to be opened appropriately by an external opener. Thus by default on Linux, drive will create \*.desktop files for files that fall into this category.

## Command Aliases

`drive` supports a few aliases to make usage familiar to the utilities in your shell e.g:
+ cp : copy
+ ls : list 
+ mv : move
+ rm : delete


## Index Prune

* index 

If you would like to fetch missing index files for files that would otherwise not need any modifications, run:

```shell
$ drive index path1 path2 path3/path3.1 # To fetch any missing indices in those paths
$ drive index --id 0CLu4lbUI9RTRM80k8EMoe5JQY2z
```

You can also fetch specific files by prefix matches
```shell
$ drive index --matches mp3 jpg
```

* prune

In case you might have deleted files remotely but never using drive, and feel like you have stale indices,
running `drive index --prune` will search your entire indices dir for index files that do not exist remotely and remove those ones

```shell
$ drive index --prune
```

* prune-and-index
To combine both operations (prune and then fetch) for indices:

```shell
$ drive index --all-ops
```

## Url

The url command prints out the url of a file. It allows you to specify multiple paths relative to root or even by id

```shell
$ drive url Photos/2015/07/Releases intros/flux
$ drive url --id  0Bz5qQkvRAeVEV0JtZl4zVUZFWWx  1Pwu8lzYc9RTPTEpwYjhRMnlSbDQ 0Cz5qUrvDBeX4RUFFbFZ5UXhKZm8
```

## Open

The open command allows for files to be opened by the default file browser, default web browser, either by path or by id for paths that exist atleast remotely

```shell
$ drive open --file-browser=false --web-browser f1/f2/f3 jamaican.mp4
$ drive open --file-browser --id 0Bz8qQkpZAeV9T1PObvs2Y3BMQEj 0Y9jtQkpXAeV9M1PObvs4Y3BNRFk
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

## Applying patches 
To  apply patches of code e.g in the midst of bug fixes, you'll just need a little bit of git fiddling.

For example to patch your code with that on remote branch patch-1, you'll need to go into the source
code directory, fetch all content from the git remote, checkout the patch branch then run the go installation: something like this.

```shell
$ cd $GOPATH/src/github.com/odeke-em/drive
$ git fetch --all
$ git checkout patch-1
$ git pull origin patch-1
$ go get github.com/odeke-em/drive/cmd/drive
```

## Why another Google Drive client?

Background sync is not just hard, it is stupid. Here are my technical and philosophical rants about why it is not worth to implement:

* Too racy. Data is shared between your remote resource, local disk and sometimes in your sync daemon's in-memory structs. Any party could touch a file at any time. It is hard to lock these actions. You end up working with multiple isolated copies of the same file and trying to determine which is the latest version that should be synced across different contexts.

* It requires great scheduling to perform best with your existing environmental constraints. On the other hand, file attribute have an impact on the sync strategy. Large files block -- you wouldn't like to sit on and wait for a VM image to get synced before you can start working on a tiny text file.

* It needs to read your mind to understand your priorities. Which file do you need most? It needs to read your mind to foresee your future actions. I'm editing a file, and saving the changes time to time. Why not to wait until I feel confident enough to commit the changes remotely?

`drive` is not a sync daemon, it provides:

* Upstreaming and downstreaming. Unlike a sync command, we provide pull and push actions. The user has the opportunity to decide what to do with their local copy and when they decide to. Make some changes, either push the file remotely or revert it to the remote version. You can perform these actions with user prompt:

	    $ echo "hello" > hello.txt
	    $ drive push # pushes hello.txt to Google Drive
	    $ echo "more text" >> hello.txt
	    $ drive pull # overwrites the local changes with the remote version

* Allowing to work with a specific file or directory, optionally not recursively. If you recently uploaded a large VM image to Google Drive, yet only a few text files are required for you to work, simply only push/pull the exact files you'd like to worth with:

	    $ echo "hello" > hello.txt
	    $ drive push hello.txt # pushes only the specified file
	    $ drive pull path/to/a/b path2/to/c/d/e # pulls the remote directory recursively

* Better I/O scheduling. One of the major goals is to provide better scheduling to improve upload/download times.

* Possibility to support multiple accounts. Pull from or push to multiple Google Drive remotes. Possibility to support multiple backends. Why not to push to Dropbox or Box as well?

## Known issues

* It probably doesn't work on Windows.
* Google Drive allows a directory to contain files/directories with the same name. Client doesn't handle these cases yet. We don't recommend you to use `drive` if you have such files/directories to avoid data loss.
* Racing conditions occur if remote is being modified while we're trying to update the file. Google Drive provides resource versioning with ETags, use Etags to avoid racy cases.
* drive rejects reading from namedPipes because they could infinitely hang. See [issue #208](https://github.com/odeke-em/drive/issues/208).


## Reach out

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
