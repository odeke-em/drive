// Copyright 2016 Google Inc. All Rights Reserved.
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
	"net/url"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

const (
	ProjectNewIssueUrl = "https://github.com/odeke-em/drive/issues/new"
)

func (g *Commands) FileIssue() error {
	title := ""
	body := ""
	newIssueRequestUrl := ProjectNewIssueUrl

	piped := false

	meta := g.opts.Meta
	if meta != nil {
		metaDeref := *meta
		titlePieces, _ := metaDeref[IssueTitleKey]
		title = sepJoin("\n", titlePieces...)

		bodyPieces, _ := metaDeref[IssueBodyKey]
		body = sepJoin("\n", bodyPieces...)

		if _, pipedSet := metaDeref[PipedKey]; pipedSet { // Intentionally verbose
			piped = true
		}
	}

	if piped {
		clauses, err := readFileFromStdin(noopOnIgnorer)
		if err != nil {
			return err
		}

		body = sepJoin("\n", clauses...)
	}

	suffixed := []string{}
	if trimmedTitle := strings.Trim(title, " "); trimmedTitle != "" {
		suffixed = append(suffixed, fmt.Sprintf("title=%s", url.QueryEscape(trimmedTitle)))
	}

	if trimmedBody := strings.Trim(body, " "); trimmedBody != "" {
		suffixed = append(suffixed, fmt.Sprintf("body=%s", url.QueryEscape(trimmedBody)))
	}

	if len(suffixed) >= 1 {
		newIssueRequestUrl += fmt.Sprintf("?%s", sepJoin("&", suffixed...))
	}

	return open.Run(newIssueRequestUrl)
}
