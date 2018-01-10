package namespace_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/odeke-em/namespace"
)

func TestParse(t *testing.T) {
	tests := [...]struct {
		text    string
		wantErr bool
		want    namespace.Namespace
	}{
		0: {
			text: `[]
				key1=value2

				[    ]
				k2=v2

				[push/pull////]
				k2=v2
				[pull]
				kp2=vp2`,
			want: namespace.Namespace{
				"global": []string{"key1=value2", "k2=v2"},
				"pull":   []string{"k2=v2", "kp2=vp2"},
				"push":   []string{"k2=v2"},
			},
		},
	}

	for i, tt := range tests {
		r := strings.NewReader(tt.text)
		ns, err := namespace.Parse(r)
		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: err=nil", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err=%v", i, err)
			continue
		}

		gotBlob, _ := json.MarshalIndent(ns, "", "  ")
		wantBlob, _ := json.MarshalIndent(tt.want, "", "  ")
		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("got: %s\nwant: %s", gotBlob, wantBlob)
		}
	}
}
