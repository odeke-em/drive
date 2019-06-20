### Platform Specific Installation Methods

---
### Arch Linux or Arch based distros.
This includes Arch linux, Antergos, Manjaro, etc. [List](https://wiki.archlinux.org/index.php/Arch_based_distributions_(active))

```sh
yay -S drive-bin
```
Since drive is in the aur, you will need an aur helper such as yay above. If you are not fimilar with
a helper, you can find a list [here](https://wiki.archlinux.org/index.php/AUR_helpers#AUR_search.2Fbuild_helpers)

---
### Debian, or Debian based distros.
You may need to install the package software-properties-common to use apt-add-repository command.

```sh
sudo apt-get install software-properties-common dirmngr
```

After installing software-properties-common, you can run these commands. Updates will be as normal with all debian packages.
Note: The apt-key command is no longer required on apt 1.1 systems. It's safe to ignore any error presented.

```sh
sudo apt-add-repository 'deb http://shaggytwodope.github.io/repo ./'
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 7086E9CC7EC3233B
sudo apt-key update
sudo apt-get update
sudo apt-get install drive
```

### Automation on the latest stable Debian distribution. 
This GitLab-hosted [DR](https://gitlab.com/jean-christophe-manciot/dr) offers:
- the latest compiled drive client
- scripts for automated drive commands including syncing
- the latest version of many other popular packages.

Replace <os_code_name> by latest Debian code name, for instance stretch (only the latest Debian distribution is supported):
```sh
sudo gpg --keyserver pgpkeys.mit.edu --recv-keys DF7396D82BBA3684FCCADD4DB063838ED13997FD
sudo bash -c 'gpg --export --armor DF7396D82BBA3684FCCADD4DB063838ED13997FD | apt-key add -'
sudo bash -c $'echo "deb https://gitlab.com/jean-christophe-manciot/dr/raw/master/Debian <os_code_name> stable #JC Manciot\'s DR" > /etc/apt/sources.list.d/jean-christophe-manciot.list'
sudo bash -c $'echo "deb https://gitlab.com/jean-christophe-manciot/dr/raw/master/Debian <os_code_name> unstable #JC Manciot\'s DR" > /etc/apt/sources.list.d/jean-christophe-manciot.list'
sudo apt-get update
sudo apt-get install drive-google
```

You may need to install Comodo RSA Domain Validation Secure Server Certificate Authority used by Gitlab. Cf. the [DR](https://gitlab.com/jean-christophe-manciot/dr) for details.

---
### Fedora distro.
Fedora rpms are available from the Fedora Copr project [here](https://copr.fedorainfracloud.org/coprs/vaughan/drive-google/)

Enable the copr repository:

```
dnf copr enable vaughan/drive-google
```

Install the package (drive-google):

```
dnf install drive-google
```

---
### openSUSE distro. (may also work with CentOS, Red Hat)
```sh
# install needed software tools
sudo yum install go mercurial git hg-git
mkdir $HOME/go
export GOPATH=$HOME/go
# For convenience, add the workspace's bin subdirectory to your PATH:
export PATH=$PATH:$GOPATH/bin

# get and compile the drive program
go get github.com/odeke-em/drive/cmd/drive

# run drive with this command:
$GOPATH/bin/drive
```

---
### Ubuntu, or Ubuntu based distros. 
This [PPA](https://launchpad.net/~twodopeshaggy/+archive/ubuntu/drive) includes Ubuntu, Mint, Linux Lite, etc. [List](http://distrowatch.com/search.php?basedon=Ubuntu)

```sh
sudo add-apt-repository ppa:twodopeshaggy/drive
sudo apt-get update
sudo apt-get install drive
```

### Automation on the latest stable Ubuntu distribution. 
This GitLab-hosted [PPA](https://gitlab.com/jean-christophe-manciot/ppa) offers:
- the latest compiled drive client
- scripts for automated drive commands including syncing
- the latest version of many other popular packages.

Replace <os_code_name> by latest Ubuntu code name, for instance artful (only the latest Ubuntu distribution is supported):
```sh
sudo gpg --keyserver keyserver.ubuntu.com --recv-keys DF7396D82BBA3684FCCADD4DB063838ED13997FD
sudo bash -c 'gpg --export --armor DF7396D82BBA3684FCCADD4DB063838ED13997FD | apt-key add -'
sudo bash -c $'echo "deb https://gitlab.com/jean-christophe-manciot/ppa/raw/master/Ubuntu <os_code_name> stable #JC Manciot\'s Stable PPA" >> /etc/apt/sources.list.d/jean-christophe-manciot.list'
sudo bash -c $'echo "deb https://gitlab.com/jean-christophe-manciot/ppa/raw/master/Ubuntu <os_code_name> unstable #JC Manciot\'s Unstable PPA" >> /etc/apt/sources.list.d/jean-christophe-manciot.list'
sudo apt-get update
sudo apt-get install drive-google
```

You may need to install Comodo RSA Domain Validation Secure Server Certificate Authority used by Gitlab. Cf. the [PPA](https://gitlab.com/jean-christophe-manciot/ppa) for details.

---
### Packages Provided By

Platform | Author |
---------| -------|
[Arch Linux](https://aur.archlinux.org/packages/drive-bin) | [Alex Dewar](https://github.com/alexdewar)
[Debian Linux](http://shaggytwodope.github.io/repo) | [Jonathan Jenkins](https://github.com/shaggytwodope)
[Automation on Debian Linux](https://gitlab.com/jean-christophe-manciot/dr) | [Jean-Christophe Manciot](https://gitlab.com/jean-christophe-manciot)
[Fedora Linux](https://copr.fedorainfracloud.org/coprs/vaughan/drive-google/) | [Vaughan Agrez](https://github.com/agrez)
[openSUSE Linux]() | [Grant Rostig](https://github.com/grantrostig)
[Ubuntu Linux](https://launchpad.net/~twodopeshaggy/+archive/ubuntu/drive) | [Jonathan Jenkins](https://github.com/shaggytwodope)
[Automation on Ubuntu Linux](https://gitlab.com/jean-christophe-manciot/ppa) | [Jean-Christophe Manciot](https://gitlab.com/jean-christophe-manciot)

