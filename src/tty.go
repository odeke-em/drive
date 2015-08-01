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

// +build cgo

package drive

/*
#include <termios.h>
#include <unistd.h>

void flush_tty_in() {
	if (isatty(0))
		tcflush(0, TCIFLUSH);
}
*/
import "C"

// flushTTYin flushes the input buffer of the tty.
// Use it before asking the user to make decisions, especially after
// a long wait.
func flushTTYin() {
	C.flush_tty_in()
}
