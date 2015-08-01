// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drive

import (
	"fmt"
)

const (
	AboutKey      = "about"
	AllKey        = "all"
	CopyKey       = "copy"
	DeleteKey     = "delete"
	DiffKey       = "diff"
	EmptyTrashKey = "emptytrash"
	FeaturesKey   = "features"
	HelpKey       = "help"
	InitKey       = "init"
	DeInitKey     = "deinit"
	LinkKey       = "Link"
	ListKey       = "list"
	MoveKey       = "move"
	OSLinuxKey    = "linux"
	PullKey       = "pull"
	PushKey       = "push"
	PubKey        = "pub"
	RenameKey     = "rename"
	QuotaKey      = "quota"
	ShareKey      = "share"
	StatKey       = "stat"
	TouchKey      = "touch"
	TrashKey      = "trash"
	UnshareKey    = "unshare"
	UntrashKey    = "untrash"
	UnpubKey      = "unpub"
	VersionKey    = "version"
	Md5sumKey     = "md5sum"
	NewKey        = "new"
	IndexKey      = "index"
	PruneKey      = "prune"

	CoercedMimeKeyKey     = "coerced-mime"
	DepthKey              = "depth"
	EmailsKey             = "emails"
	EmailMessageKey       = "emailMessage"
	ForceKey              = "force"
	QuietKey              = "quiet"
	QuitShortKey          = "q"
	YesShortKey           = "Y"
	QuitLongKey           = "quit"
	MatchesKey            = "matches"
	HiddenKey             = "hidden"
	Md5Key                = "md5"
	NoPromptKey           = "no-prompt"
	SizeKey               = "size"
	NameKey               = "name"
	OriginalNameKey       = "oname"
	ModTimeKey            = "modt"
	LastViewedByMeTimeKey = "lvt"
	RoleKey               = "role"
	TypeKey               = "type"
	TrashedKey            = "trashed"
	SkipMimeKeyKey        = "skip-mime"
	MatchMimeKeyKey       = "exact-mime"
	ExactTitleKey         = "exact-title"
	MatchOwnerKey         = "match-owner"
	ExactOwnerKey         = "exact-owner"
	NotOwnerKey           = "skip-owner"
	SortKey               = "sort"
	FolderKey             = "folder"
	MimeKey               = "mime-key"
	DriveRepoRelPath      = "github.com/odeke-em/drive"
)

const (
	DescAbout                 = "print out information about your Google drive"
	DescAll                   = "print out the entire help section"
	DescCopy                  = "copy remote paths to a destination"
	DescDelete                = "deletes the items permanently. This operation is irreversible"
	DescDiff                  = "compares local files with their remote equivalent"
	DescEmptyTrash            = "permanently cleans out your trash"
	DescExcludeOps            = "exclude operations"
	DescFeatures              = "returns information about the features of your drive"
	DescIndex                 = "fetch indices from remote"
	DescHelp                  = "Get help for a topic"
	DescInit                  = "initializes a directory and authenticates user"
	DescDeInit                = "removes the user's credentials and initialized files"
	DescList                  = "lists the contents of remote path"
	DescMove                  = "move files/folders"
	DescQuota                 = "prints out information related to your quota space"
	DescPublish               = "publishes a file and prints its publicly available url"
	DescRename                = "renames a file/folder"
	DescPull                  = "pulls remote changes from Google Drive"
	DescPruneIndices          = "remove stale indices"
	DescPush                  = "push local changes to Google Drive"
	DescShare                 = "share files with specific emails giving the specified users specifies roles and permissions"
	DescStat                  = "display information about a file"
	DescTouch                 = "updates a remote file's modification time to that currently on the server"
	DescTrash                 = "moves files to trash"
	DescUnshare               = "revoke a user's access to a file"
	DescUntrash               = "restores files from trash to their original locations"
	DescUnpublish             = "revokes public access to a file"
	DescVersion               = "prints the version"
	DescMd5sum                = "prints a list compatible with md5sum(1)"
	DescAccountTypes          = "\n\t* anyone.\n\t* user.\n\t* domain.\n\t* group"
	DescRoles                 = "\n\t* owner.\n\t* reader.\n\t* writer.\n\t* commenter."
	DescExplicitylPullExports = "explicitly pull exports"
	DescIgnoreChecksum        = "avoids computation of checksums as a final check." +
		"\nUse cases may include:\n\t* when you are low on bandwidth e.g SSHFS." +
		"\n\t* Are on a low power device"
	DescIgnoreConflict     = "turns off the conflict resolution safety"
	DescIgnoreNameClashes  = "ignore name clashes"
	DescSort               = "sort items in the order\n\t* md5.\n\t* name.\n\t* size.\n\t* type.\n\t* version"
	DescSkipMime           = "skip elements with mimeTypes derived from these extensison"
	DescMatchMime          = "get elements with the exact mimeTypes derived from extensisons"
	DescMatchTitle         = "elements with matching titles"
	DescExactTitle         = "get elements with the exact titles"
	DescMatchOwner         = "elements with matching owners"
	DescExactOwner         = "elements with the exact owner"
	DescNotOwner           = "ignore elements owned by these users"
	DescNew                = "create a new file/folder"
	DescAllIndexOperations = "perform all the index related operations"
)

const (
	CLIOptionExplicitlyExport   = "explicitly-export"
	CLIOptionIgnoreChecksum     = "ignore-checksum"
	CLIOptionIgnoreConflict     = "ignore-conflict"
	CLIOptionIgnoreNameClashes  = "ignore-name-clashes"
	CLIOptionExcludeOperations  = "exclude-ops"
	CLIOptionId                 = "id"
	CLIOptionNoClobber          = "no-clobber"
	CLIOptionNotify             = "notify"
	CLIOptionSkipMime           = "skip-mime"
	CLIOptionMatchMime          = "exact-mime"
	CLIOptionExactTitle         = "exact-title"
	CLIOptionMatchTitle         = "match-mime"
	CLIOptionExactOwner         = "exact-owner"
	CLIOptionMatchOwner         = "match-owner"
	CLIOptionNotOwner           = "skip-owner"
	CLIOptionPruneIndices       = "prune"
	CLIOptionAllIndexOperations = "all-ops"
)

const (
	GoogleApiClientIdEnvKey     = "GOOGLE_API_CLIENT_ID"
	GoogleApiClientSecretEnvKey = "GOOGLE_API_CLIENT_SECRET"
)

const (
	DesktopExtension = "desktop"
)

const (
	InfiniteDepth = -1
)

var skipChecksumNote = fmt.Sprintf(
	"\nNote: You can skip checksum verification by passing in flag `-%s`", CLIOptionIgnoreChecksum)

var docMap = map[string][]string{
	AboutKey: []string{
		DescAbout,
	},
	CopyKey: []string{
		DescCopy,
	},
	DeleteKey: []string{
		DescDelete,
	},
	DiffKey: []string{
		DescDiff, "Accepts multiple remote paths for line by line comparison",
		skipChecksumNote,
	},
	EmptyTrashKey: []string{
		DescEmptyTrash,
	},
	FeaturesKey: []string{
		DescFeatures,
	},
	InitKey: []string{
		DescInit, "Requests for access to your Google Drive",
		"Creating a folder that contains your credentials",
		"Note: `init` in an already initialized drive will erase the old credentials",
	},
	PullKey: []string{
		DescPull, "Downloads content from the remote drive or modifies",
		" local content to match that on your Google Drive",
		skipChecksumNote,
	},
	PushKey: []string{
		DescPush, "Uploads content to your Google Drive from your local path",
		"Push comes in a couple of flavors",
		"\t* Ordinary push: `drive push path1 path2 path3`",
		"\t* Mounted push: `drive push -m path1 [path2 path3] drive_context_path`",
		skipChecksumNote,
	},
	ListKey: []string{
		DescList,
		"List the information of a remote path not necessarily present locally",
		"Allows printing of long options and by default does minimal printing",
	},
	MoveKey: []string{
		DescMove,
		"Moves files/folders between folders",
	},
	PubKey: []string{
		DescPublish, "Accepts multiple paths",
	},
	RenameKey: []string{
		DescRename, "Accepts <src> <newName>",
	},
	QuotaKey: []string{DescQuota},
	ShareKey: []string{
		DescShare, "Accepts multiple paths",
		"Specify the emails to share with as well as the message to send them on notification",
		"Accepted values for:\n+ accountType: ",
		DescAccountTypes, "\n+ roles:", DescRoles,
	},
	StatKey: []string{
		DescStat, "provides detailed information about a remote file",
		"Accepts multiple paths",
	},
	TouchKey: []string{
		DescTouch, "Given a list of remote files `touch` updates their",
		"last edit times to that currently on the server",
	},
	TrashKey: []string{
		DescTrash, "Sends a list of remote files to trash",
	},
	UnshareKey: []string{
		DescUnshare, "Accepts multiple paths",
		"Accepted values for accountTypes::", DescAccountTypes,
	},
	UntrashKey: []string{
		DescUntrash, "takes remote files out of the trash",
		"Note: untrash is a relative path command so any resolutions are made",
		"relative to the current working directory i.e",
		"\n\t$ drive trash mnt/logos",
	},
	UnpubKey: []string{
		DescUnpublish, "revokes public access to a list of remote files",
	},
	VersionKey: []string{
		DescVersion, fmt.Sprintf("current version is: %s", Version),
	},
}

var Aliases = map[string][]string{
	CopyKey: []string{"cp"},
	ListKey: []string{"ls"},
	MoveKey: []string{"mv"},
}

func ShowAllDescriptions() {
	for key, _ := range docMap {
		ShowDescription(key)
		fmt.Println()
	}
}

func ShowDescriptions(topics ...string) {
	if len(topics) < 1 {
		topics = append(topics, AllKey)
	}

	for _, topic := range topics {
		ShowDescription(topic)
		fmt.Println()
	}
}

func ShowDescription(topic string) {
	if topic == AllKey {
		ShowAllDescriptions()
		return
	}

	help, ok := docMap[topic]
	if !ok {
		fmt.Printf("Unkown command '%s' type `drive help all` for entire usage documentation\n", topic)
		ShowAllDescriptions()
	} else {
		description, documentation := help[0], help[1:]
		fmt.Printf("Name\n\t%s - %s\n", topic, description)
		if len(documentation) >= 1 {
			fmt.Println("Description")
			for _, line := range documentation {
				fmt.Printf("\t%s\n", line)
			}
			fmt.Printf("\n* For usage flags: \033[32m`drive %s -h`\033[00m\n\n", topic)
		}
	}
}
