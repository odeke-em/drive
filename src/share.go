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
	"strings"
	"sync"

	"github.com/odeke-em/log"
)

type AccountType int

const (
	UnknownAccountType = 1 << iota
	Anyone
	User
	Domain
	Group
)

type Role int

const (
	UnknownRole = 1 << iota
	Owner
	Reader
	Writer
	Commenter
)

const (
	NoopOnShare = 1 << iota
	Notify
)

type shareChange struct {
	emailMessage string
	emails       []string
	role         Role
	accountType  AccountType
	files        []*File
	revoke       bool
	notify       bool
}

type permission struct {
	fileId      string
	value       string
	message     string
	role        Role
	accountType AccountType
	notify      bool
}

func (r *Role) String() string {
	switch *r {
	case Owner:
		return "owner"
	case Reader:
		return "reader"
	case Writer:
		return "writer"
	case Commenter:
		return "commenter"
	}
	return "unknown"
}

func (a *AccountType) String() string {
	switch *a {
	case Anyone:
		return "anyone"
	case User:
		return "user"
	case Domain:
		return "domain"
	case Group:
		return "group"
	}
	return "unknown"
}

func stringToRole() func(string) Role {
	roleMap := make(map[string]Role)
	roles := []Role{UnknownRole, Owner, Reader, Writer, Commenter}
	for _, role := range roles {
		roleMap[role.String()] = role
	}
	return func(s string) Role {
		r, ok := roleMap[strings.ToLower(s)]
		if !ok {
			return Reader
		}
		return r
	}
}

func stringToAccountType() func(string) AccountType {
	accountMap := make(map[string]AccountType)
	accounts := []AccountType{UnknownAccountType, Anyone, User, Domain, Group}
	for _, account := range accounts {
		accountMap[account.String()] = account
	}
	return func(s string) AccountType {
		a, ok := accountMap[strings.ToLower(s)]
		if !ok {
			return User
		}
		return a
	}
}

var reverseRoleResolve = stringToRole()
var reverseAccountTypeResolve = stringToAccountType()

func (g *Commands) resolveRemotePaths(relToRootPaths []string, byId bool) (files []*File) {
	resolver := g.rem.FindByPath
	if byId {
		resolver = g.rem.FindByIdMulti
	}

	done := make(chan bool)
	doneCount := uint64(0)

	for _, relToRoot := range relToRootPaths {
		doneCount += 1

		go func(p string) {
			childFiles, err := resolver(p)
			if err == nil {
				files = append(files, childFiles...)
			}

			done <- true
		}(relToRoot)
	}

	for i := uint64(0); i < doneCount; i++ {
		<-done
	}

	return files
}

func emailsToIds(g *Commands, emails []string) map[string]string {
	emailToIds := make(map[string]string)
	var wg sync.WaitGroup
	wg.Add(len(emails))
	for _, email := range emails {
		go func(email string, wgg *sync.WaitGroup) {
			defer wgg.Done()
			emailId, err := g.rem.idForEmail(email)
			if err == nil {
				emailToIds[email] = emailId
			}
		}(email, &wg)
	}
	wg.Wait()
	return emailToIds
}

func (c *Commands) Unshare(byId bool) (err error) {
	return c.share(true, byId)
}

func (c *Commands) Share(byId bool) (err error) {
	return c.share(false, byId)
}

func showPromptShareChanges(logy *log.Logger, change *shareChange) bool {
	if len(change.files) < 1 {
		return false
	}
	if change.revoke {
		logy.Logf("Revoke access for accountType: \033[92m%s\033[00m for file(s):\n",
			change.accountType.String())
		for _, file := range change.files {
			logy.Logln("+ ", file.Name)
		}
		logy.Logln()
		return promptForChanges()
	}

	if len(change.emails) < 1 {
		return false
	}

	if change.notify {
		logy.Logln("Message:\n\t", change.emailMessage)
	}

	logy.Logln("Receipients:")
	for _, email := range change.emails {
		logy.Logf("\t\033[92m+\033[00m %s\n", email)
	}

	logy.Logln("\nFile(s) to share:")
	for _, file := range change.files {
		if file == nil {
			continue
		}
		logy.Logf("\t\033[92m+\033[00m %s\n", file.Name)
	}
	return promptForChanges()
}

func (c *Commands) playShareChanges(change *shareChange) error {
	if c.opts.canPrompt() && !showPromptShareChanges(c.log, change) {
		return nil
	}

	for _, file := range change.files {
		if change.revoke {
			if err := c.rem.deletePermissions(file.Id, change.accountType); err != nil {
				return fmt.Errorf("%s: %v", file.Name, err)
			}
			continue
		}

		for _, email := range change.emails {
			perm := permission{
				fileId:      file.Id,
				value:       email,
				message:     change.emailMessage,
				notify:      change.notify,
				role:        change.role,
				accountType: change.accountType,
			}
			_, err := c.rem.insertPermissions(&perm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Commands) share(revoke, byId bool) (err error) {
	files := c.resolveRemotePaths(c.opts.Sources, byId)

	var role Role
	var accountType AccountType
	var emails []string
	var emailMessage string

	// Setup the defaults
	role = Reader
	accountType = User
	meta := *c.opts.Meta

	if meta != nil {
		emailList, eOk := meta["emails"]
		if eOk {
			emails = emailList
			if false {
				emailIdMap := emailsToIds(c, emailList)
				c.log.Logln(emailIdMap)
			}
		}

		roleList, rOk := meta["role"]
		if rOk && len(roleList) >= 1 {
			role = reverseRoleResolve(roleList[0])
		}
		accountTypeList, aOk := meta["accountType"]
		if aOk && len(accountTypeList) >= 1 {
			accountType = reverseAccountTypeResolve(accountTypeList[0])
		}

		emailMessageList, emOk := meta["emailMessage"]
		if emOk && len(emailMessageList) >= 1 {
			emailMessage = strings.Join(emailMessageList, "\n")
		}
	}

	notify := (c.opts.TypeMask & Notify) != 0

	change := shareChange{
		accountType:  accountType,
		emailMessage: emailMessage,
		emails:       emails,
		files:        files,
		revoke:       revoke,
		role:         role,
		notify:       notify,
	}

	return c.playShareChanges(&change)
}
