package tmpfile

import (
	"os"
	"reflect"
	"testing"
)

func TestTmpFile(t *testing.T) {
	testCases := [...]struct {
		ctx     *Context
		wantErr bool
	}{
		0: {ctx: nil, wantErr: true},
		1: {ctx: &Context{}},
		2: {ctx: &Context{
			CreateIsolatedDir: true,
			Suffix:            "tmpf",
		}},
		3: {ctx: &Context{
			CreateIsolatedDir: true,
			Suffix:            "",
		}},
		4: {ctx: &Context{
			Dir:    "here",
			Suffix: "",
		}},
		5: {ctx: &Context{
			Dir:    "./here",
			Suffix: "",
		}},
		6: {ctx: &Context{
			Dir:    " ",
			Suffix: "",
		}},
		7: {ctx: &Context{
			Dir:    "..",
			Suffix: "",
		}},
		8: {ctx: &Context{
			Dir:               "tx1",
			Suffix:            "",
			CreateIsolatedDir: true,
		}},
	}

	for i, tt := range testCases {
		tmpf, err := New(tt.ctx)

		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: got nil err, want non-nil err", i)
			}
		} else {
			if err != nil {
				t.Fatalf("#%d: got err=%v, want nil err", i, err)
			}
			if tmpf == nil {
				t.Fatalf("#%d: got nil tempfile, want non-nil tempfile", i)
			}

			externStatInfo, err := os.Stat(tmpf.Name())
			if err != nil {
				t.Errorf("#%d: got statErr=%v, expected nil stat error", i, err)
			} else {
				gotSys := externStatInfo.Sys()
				iStatInfo, err := tmpf.Stat()
				if err != nil {
					t.Fatalf("#%d: tmpf.Stat() err=%v, expected nil error", i, err)
				}
				iSys := iStatInfo.Sys()
				if !reflect.DeepEqual(iSys, gotSys) {
					t.Errorf("#%d: tmpfSys=%q os.StatByNameSys=%q", i, iSys, gotSys)
				}
			}
		}

		if tmpf != nil {
			// Invoke done more than once to ensure it fires only once
			for j := 0; j < 4; j++ {
				if err := tmpf.Done(); err != nil {
					t.Fatalf("#%d, Done.Try #%d: done.Err=%v expected non-nil error", i, j, err)
				}
			}

			if tt.ctx.CreateIsolatedDir {
				// Ensure that deletions were performed properly when
				dir := tmpf.Dir()
				if dir == "" {
					t.Fatalf("#%d: got=%q, want non empty dir", i, dir)
				}
				dirStatInfo, err := os.Stat(dir)
				if dirStatInfo != nil {
					t.Errorf("#%d: dirStatInfo=%v expected nil", i, dirStatInfo)
				}
				if !os.IsNotExist(err) {
					t.Errorf("#%d: got err=%v, expected one of os.IsNotExist", i, err)
				}
			}
		}
	}
}
