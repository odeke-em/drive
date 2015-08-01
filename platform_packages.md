### Platform Specific Installation Methods


### Arch Linux or Arch based distros.
This includes Arch linux, Antergos, Manjaro, etc. [List](https://wiki.archlinux.org/index.php/Arch_based_distributions_(active))

```sh
$ yaourt -S drive
```
Since drive is in the aur, you will need an aur helper such as yaourt above. If you are not fimilar with
a helper, you can find a list [here](https://wiki.archlinux.org/index.php/AUR_helpers#AUR_search.2Fbuild_helpers)


### Ubuntu, or Ubuntu based distros. 
This includes Ubuntu, Mint, Linux Lite, etc. [List](http://distrowatch.com/search.php?basedon=Ubuntu)

```sh
$ sudo add-apt-repository ppa:twodopeshaggy/drive
$ sudo apt-get update
$ sudo apt-get install drive
```

### openSUSE distro. (may also work with fedora, CentOS, Red Hat)
```sh
# install needed software tools
sudo yum install go mercurial git hg-git
$ mkdir $HOME/go
$ export GOPATH=$HOME/go
# For convenience, add the workspace's bin subdirectory to your PATH:
$ export PATH=$PATH:$GOPATH/bin

# get and compile the drive program
$ go get github.com/odeke-em/drive/cmd/drive

# run drive with this command:
$ $GOPATH/bin/drive
```
### Packages Provided By

Platform | Author |
---------| -------|
[Arch Linux](https://aur.archlinux.org/packages/drive) | [Jonathan Jenkins](https://github.com/shaggytwodope)
[Ubuntu Linux](https://launchpad.net/~twodopeshaggy/+archive/ubuntu/drive) | [Jonathan Jenkins](https://github.com/shaggytwodope)
[openSUSE Linux]() | [Grant Rostig](https://github.com/grantrostig)

