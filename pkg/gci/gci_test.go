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
		desc string
		//
		imports   string
		localFlag string
		//
		want *pkg
	}{
		{
			desc: "basic",
			//
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
				comments: map[string][]string{},
				alias:    map[string]string{},
			},
		},
		{
			desc: "same line comments",
			//
			imports: `
	"fmt" // same line comment
	"log" //nolint

	_ "database/sql"   // import sql
	_ "net/http/pprof" //nolint:golint

	"github.com/owner/repo"
`,
			want: &pkg{
				list: map[int][]string{
					standard: {`"net/http/pprof"`, `"database/sql"`, `"log"`, `"fmt"`},
					remote:   {`"github.com/owner/repo"`},
				},
				comments: map[string][]string{
					`"fmt"`:            {"// same line comment"},
					`"log"`:            {"//nolint"},
					`"database/sql"`:   {"// import sql"},
					`"net/http/pprof"`: {"//nolint:golint"},
				},
				alias: map[string]string{
					`"database/sql"`:   "_",
					`"net/http/pprof"`: "_",
				},
			},
		},
		{
			desc: "one line comments",
			imports: `
	// import sql
	_ "database/sql"
	//nolint
	"log"

	"github.com/owner/repo"
`,
			want: &pkg{
				list: map[int][]string{
					standard: {`"log"`, `"database/sql"`},
					remote:   {`"github.com/owner/repo"`},
				},
				comments: map[string][]string{
					`"log"`:          {"//nolint"},
					`"database/sql"`: {"// import sql"},
				},
				alias: map[string]string{
					`"database/sql"`: "_",
				},
			},
		},
		{
			desc: "multi line comments",
			imports: `
	// import
	// sql
	_ "database/sql"
	// Import log
	//nolint
	"log"
	// First dangling comment

	// Second dangling comment
`,
			want: &pkg{
				list: map[int][]string{
					standard: {`"log"`, `"database/sql"`},
				},
				comments: map[string][]string{
					`"log"`:          {"//nolint", "// Import log"},
					`"database/sql"`: {"// sql", "// import"},
				},
				alias: map[string]string{
					`"database/sql"`: "_",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			data := make([][]byte, 0, len(tt.imports))
			for _, line := range strings.Split(tt.imports, linebreak) {
				data = append(data, []byte(line))
			}

			pkg := newPkg(data, tt.localFlag)
			require.Equal(t, tt.want, pkg)
		})
	}
}
