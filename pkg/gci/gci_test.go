package gci

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPkgType(t *testing.T) {
	testCases := []struct {
		Line           string
		LocalFlag      string
		ExpectedResult int
	}{
		{Line: `"foo/pkg/bar"`, LocalFlag: "", ExpectedResult: remote},
		{Line: `"foo/pkg/bar"`, LocalFlag: "foo", ExpectedResult: local},
		{Line: `"foo/pkg/bar"`, LocalFlag: "bar", ExpectedResult: remote},
		{Line: `"foo/pkg/bar"`, LocalFlag: "github.com/foo/bar", ExpectedResult: remote},

		{Line: `"github.com/foo/bar"`, LocalFlag: "", ExpectedResult: remote},
		{Line: `"github.com/foo/bar"`, LocalFlag: "foo", ExpectedResult: remote},
		{Line: `"github.com/foo/bar"`, LocalFlag: "bar", ExpectedResult: remote},
		{Line: `"github.com/foo/bar"`, LocalFlag: "github.com/foo/bar", ExpectedResult: local},

		{Line: `"context"`, LocalFlag: "", ExpectedResult: standard},
		{Line: `"context"`, LocalFlag: "context", ExpectedResult: local},
		{Line: `"context"`, LocalFlag: "foo", ExpectedResult: standard},
		{Line: `"context"`, LocalFlag: "bar", ExpectedResult: standard},
		{Line: `"context"`, LocalFlag: "github.com/foo/bar", ExpectedResult: standard},

		{Line: `"os/signal"`, LocalFlag: "", ExpectedResult: standard},
		{Line: `"os/signal"`, LocalFlag: "os/signal", ExpectedResult: local},
		{Line: `"os/signal"`, LocalFlag: "foo", ExpectedResult: standard},
		{Line: `"os/signal"`, LocalFlag: "bar", ExpectedResult: standard},
		{Line: `"os/signal"`, LocalFlag: "github.com/foo/bar", ExpectedResult: standard},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s:%s", tc.Line, tc.LocalFlag), func(t *testing.T) {
			t.Parallel()

			result := getPkgType(tc.Line, tc.LocalFlag)
			if got, want := result, tc.ExpectedResult; got != want {
				t.Errorf("bad result: %d, expected: %d", got, want)
			}
		})
	}
}

func TestNewPkg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		imports   string
		localFlag string
		//
		want *pkg
	}{
		{
			imports: `
	"fmt"
	"os"

	"github.com/owner/repo"
`,
			want: &pkg{
				list: map[int][]string{
					standard: {`"os"`, `"fmt"`},
					remote:   {`"github.com/owner/repo"`},
				},
				comment: map[string]string{},
				alias:   map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			data := make([][]byte, 0, len(tt.imports))
			for _, line := range strings.Split(tt.imports, linebreak) {
				data = append(data, []byte(line))
			}

			pkg := newPkg(data, tt.localFlag)
			require.Equal(t, tt.want, pkg)
		})
	}
}
