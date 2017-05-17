#Automation Scripts

## Table of Contents

- [Services](#services)
- [Installing The Automation Scripts](#installing-the-automation-scripts)
  - [Stable Channel](#stable-channel)
  - [Unstable Channel](#unstable-channel)
- [Initializing Drive](#initializing-drive)
- [Using The Automation Scripts](#using-the-automation-scripts)
  - [Setting Up](#setting-up)
  - [Selecting Objects](#selecting-objects)
  - [Launching A Script](#launching-a-script)
  - [Help](#help)
- [drive-menu Screenshots](#drive-menu-screenshots)
- [Tracking And Filing Issues](#tracking-and-filing-issues)
- [Reaching Out](#reaching-out)
- [Disclaimer](#disclaimer)
- [LICENSE](#license)

## Services

JC Manciot's GitLab-hosted [PPA](https://gitlab.com/jean-christophe-manciot/ppa) offers a collection of scripts for automated drive commands including syncing which can be used on the latest stable Ubuntu distribution.
They are all included in the **drive-google** package.

The ```drive-<command>``` scripts offer the following services:
  + **abstraction** from underlying drive program, allowing the user to use it without knowing all its options & details
  + **automation**, allowing the user to launch each command once on all its selected objects
  + **synchronization** offering 3 types of operation:
    - *Fully-Automated*: it does not ask for user confirmation (not implemented yet)
    - *Semi-Automated*: it asks for user confirmation
    - *Manual*: it asks for user confirmation & offers a choice of a combination of 8 commands 

    The first 2 operations allow 3 different types of **master**:
    - *None*: all local & remote modified objects are synced, but trashed/deleted objects are not synced
    - *Local*: your remote Google Drive is mirrored from your local GD
    - *Remote*: your local Google Drive is mirrored from your remote GD
  + **choice of user interface**:
    - *CLI commands*: all scripts begin with ```# drive-<command>```
    - *GUI menus*: all previous commands are accessible through one command in basic graphical mode, ```# drive-menu```
  + results in basic **graphical windows**
  + **agentless**: no daemon is running on your local host

## Installing The Automation Scripts

### Stable Channel

Installing the scripts through a **stable channel** where <latest_release_code_name> below is the first word only of the latest Ubuntu release (yakkety for example):
```sh
# sudo sh -c $'echo "deb https://gitlab.com/jean-christophe-manciot/ppa/raw/master/Ubuntu <latest_release_code_name> stable #JC Manciot\'s Stable PPA" >> /etc/apt/sources.list.d/jean-christophe-manciot.list'
# sudo apt-get update
# sudo apt-get install drive-google
```

### Unstable Channel

Installing the scripts through an **unstable channel** where <latest_release_code_name> below is the first word only of the latest Ubuntu release (yakkety for example):
```sh
sudo sh -c $'echo "deb https://gitlab.com/jean-christophe-manciot/ppa/raw/master/Ubuntu <latest_release_code_name> unstable #JC Manciot\'s Unstable PPA" >> /etc/apt/sources.list.d/jean-christophe-manciot.list'
```

## Initializing Drive
```sh
# mkdir GD
# cd GD
GD# drive init
Visit this URL to get an authorization code
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=354790962074-7rrlnuanmamgg1i4feed12dpuq871bvd.apps.googleusercontent.com&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive&state=2017-01-10+08%3A03%3A00.659412311+%2B0100+CET2596996162
Paste the authorization code: *****************
GD# cp -v /etc/drive-google/.* .
cp: omitting directory '/etc/drive-google/.'
cp: omitting directory '/etc/drive-google/..'
'/etc/drive-google/.driveignore' -> './.driveignore'
'/etc/drive-google/.driverc' -> './.driverc'
'/etc/drive-google/.driveselect' -> './.driveselect'
```

## Using The Automation Scripts

### Setting Up

The general settings described in ```<google_drive_folder>/.driverc``` control how the ```drive-<command>``` scripts behave for all commands using ```[global]``` section, or for a specific command under its ```[<command>]``` section which overrides the global settings only for that command.
It's possible to group multiple settings for a bundle of commands under a ```[<command1>/<command2>/...]``` section.

More details in [Configuring General Settings](https://github.com/odeke-em/drive#configuring-general-settings) and with ```more /etc/drive-google/.driverc```.

### Selecting Objects

By default, each drive command operates over all objects (folders/files) which are located below the current working directory, as far as recursively possible (depth=-1).

It is possible to alter this behavior through the following settings:

Settings | File Location
-------- | -------------
Controlling the [traversal depth](https://github.com/odeke-em/drive#traversal-depth) under the right ```[<command>]``` section | ```<google_drive_folder>/.driverc```
Excluding/including objects with **regular expressions** | ```<google_drive_folder>/.driveignore```
Including/excluding objects with **standard wildcards** under the right ```[<command>]``` section | ```./.driveselect```

Regarding **new .driveselect**:
- It is intended for users who are not familiar with regular expressions. 
- It is possible to place .driveselect in any directory below your ```<google_drive_folder>```. The ```drive-<command>``` script selects the one located in the directory where the call has been made, allowing you to limit the length of the filenames: their path is relative to the CLI location. This is useful when the tree is very deep and there is a need to include/exclude files which are located very far in the depth of the tree. Thus, the filenames are preceded with a minimal or non existent path.
- It is also possible to list objects per type of operation or for all operations.  
- The selection logic is inverted as compared to .driveignore:
  + folders/files which are listed are included
  + folders/files which are preceded with a ! are excluded

More details are available in:
- [Configuring General Settings](https://github.com/odeke-em/drive#configuring-general-settings) and ``` # more /etc/drive-google/.driverc```
- [Excluding And Including Objects](https://github.com/odeke-em/drive#excluding-and-including-objects) and ``` # more /etc/drive-google/.driveignore```
- Including And Excluding Objects with ``` # more /etc/drive-google/.driveselect```

### Launching A Script

Once the general settings are configured and affected objects are selected, major drive commands can be automatically launched in the folder of your choice with:
```
# cd <folder>
# drive-menu
or
# drive-<command>
```

- No option is necessary on the CLI.
- If you launch a script outside a Google Drive folder, it will locate the first one it finds and change the working directory to the latter.
- ```# drive-<tabulation>``` lists all available scripts.
- ```# drive-menu``` launches a basic GUI where all commands are accessible, including the abilities to change the current working directory and to configure the general settings & selected objects without exiting drive-menu.
- ```# drive-sync``` offers syncing services not available with current drive. All details are available with:
  + ```# drive-sync --help```
  + ```# drive-menu --> Help --> Sync...```

The following commands are available within **drive-menu**:

GUI Features | CLI Equivalents
------------ | ---------------
About Drive | ```# drive-about```
Check selected remote duplicated objects | ```# drive-check-duplicates``` (clashes)
Change working directory | ```# cd <folder>```
Delete selected remote objects | ```# drive-delete```
Disable local Google Drive | ```# drive-disable```
Empty remote trash | ```# drive-empty-trash```
Help | ```# drive-<command> --help```
List selected remote objects | ```# drive-list```
List selected shared remote objects | ```# drive-list-shared```
List selected starred remote objects | ```# drive-list-starred```
Publish selected remote objects | ```# drive-publish```
Pull selected remote objects | ```# drive-pull```
Push selected local objects | ```# drive-push```
Query statistics about selected remote objects | ```# drive-stats``` (stat)
Select objects & configure general settings | Edit configuration files
Share with link selected remote objects | ```# drive-share-link```
Sync selected local/remote objects | ```# drive-sync```
Touch selected remote objects | ```# drive-touch```
Trash selected remote objects | ```# drive-trash```
Unpublish selected remote objects | ```# drive-unpublish```
Unshare selected remote objects | ```# drive-unshare```
Untrash selected remote objects | ```# drive-untrash```
Versions | ```# drive-version```

### Help

Detailed Help is available for all commands with:
- ```# drive-<command> --help```
- ```# drive-menu --> Help --> <Command>```

## drive-menu Screenshots

Some **screenshots** are available in the [automation scripts wiki](https://gitlab.com/jean-christophe-manciot/Drive/wikis/home)

## Tracking And Filing Issues

This can be done [here](https://gitlab.com/jean-christophe-manciot/Drive/issues).

## Reaching Out

Doing anything interesting with ```drive-<command>``` scripts or ready to share your favorite tips and tricks? 
Check out the [drive scripts wiki](https://gitlab.com/jean-christophe-manciot/Drive/wikis/home) and feel free to reach out with ideas for features or requests.

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
