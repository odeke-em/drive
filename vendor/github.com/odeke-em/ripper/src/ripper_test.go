// Copyright 2015 Emmanuel Odeke. All Rights Reserved.
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

package ripper

import (
	"fmt"
	"testing"
)

func TestApacheLicenseRip(t *testing.T) {
	filename := CallerFilename()

	content, err := ApacheTopLicenseRip(filename)
	if err != nil {
		t.Errorf("not expecting an error: got %v", err)
	}
	fmt.Println(content)
}
