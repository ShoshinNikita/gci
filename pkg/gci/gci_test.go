package gci

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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
				comments: map[string][]importComment{},
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
				comments: map[string][]importComment{
					`"fmt"`:            {{comment: "// same line comment", sameLine: true}},
					`"log"`:            {{comment: "//nolint", sameLine: true}},
					`"database/sql"`:   {{comment: "// import sql", sameLine: true}},
					`"net/http/pprof"`: {{comment: "//nolint:golint", sameLine: true}},
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
				comments: map[string][]importComment{
					`"log"`:          {{comment: "//nolint", sameLine: false}},
					`"database/sql"`: {{comment: "// import sql", sameLine: false}},
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
				comments: map[string][]importComment{
					`"log"`: {
						{comment: "//nolint", sameLine: false},
						{comment: "// Import log", sameLine: false},
					},
					`"database/sql"`: {
						{comment: "// sql", sameLine: false},
						{comment: "// import", sameLine: false},
					},
				},
				alias: map[string]string{
					`"database/sql"`: "_",
				},
			},
		},
		{
			desc: "mixed comments",
			imports: `
	// import
	// sql
	"database/sql" //nolint:golint
`,
			want: &pkg{
				list: map[int][]string{
					standard: {`"database/sql"`},
				},
				comments: map[string][]importComment{
					`"database/sql"`: {
						{comment: "//nolint:golint", sameLine: true},
						{comment: "// sql", sameLine: false},
						{comment: "// import", sameLine: false},
					},
				},
				alias: map[string]string{},
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

func TestFmtPkg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pkg *pkg
		//
		want string
	}{
		{
			pkg: &pkg{
				list: map[int][]string{
					standard: {`"os"`, `"fmt"`},
					remote:   {`"github.com/owner/repo"`},
				},
			},
			//
			want: `
	"fmt"
	"os"

	"github.com/owner/repo"
`,
		},
		{
			pkg: &pkg{
				list: map[int][]string{
					standard: {`"net/http/pprof"`, `"database/sql"`, `"log"`, `"fmt"`},
					remote:   {`"github.com/owner/repo"`},
				},
				comments: map[string][]importComment{
					`"fmt"`:            {{comment: "// same line comment", sameLine: true}},
					`"log"`:            {{comment: "//nolint", sameLine: true}},
					`"database/sql"`:   {{comment: "// import sql", sameLine: true}},
					`"net/http/pprof"`: {{comment: "//nolint:golint", sameLine: true}},
				},
				alias: map[string]string{
					`"database/sql"`:   "_",
					`"net/http/pprof"`: "_",
				},
			},
			//
			want: `
	_ "database/sql" // import sql
	"fmt" // same line comment
	"log" //nolint
	_ "net/http/pprof" //nolint:golint

	"github.com/owner/repo"
`,
		},
		{
			pkg: &pkg{
				list: map[int][]string{
					standard: {`"log"`, `"database/sql"`},
				},
				comments: map[string][]importComment{
					`"log"`: {
						{comment: "//nolint", sameLine: false},
						{comment: "// Import log", sameLine: false},
					},
					`"database/sql"`: {
						{comment: "// sql", sameLine: false},
						{comment: "// import", sameLine: false},
					},
				},
				alias: map[string]string{
					`"database/sql"`: "_",
				},
			},
			//
			want: `
	// import
	// sql
	_ "database/sql"
	// Import log
	//nolint
	"log"
`,
		},
		{
			pkg: &pkg{
				list: map[int][]string{
					standard: {`"database/sql"`, `"fmt"`},
					remote:   {`"github.com/remote/repo"`},
					local:    {`"github.com/local/repo"`},
				},
				comments: map[string][]importComment{
					// standart
					`"database/sql"`: {
						{comment: "//nolint:golint", sameLine: true},
						{comment: "// sql", sameLine: false},
						{comment: "// import", sameLine: false},
					},
					`"fmt"`: {
						{comment: "// fmt package", sameLine: false},
					},
					// remote
					`"github.com/remote/repo"`: {
						{comment: "//nolint", sameLine: true},
						{comment: "// test", sameLine: false},
					},
					// local
					`"github.com/local/repo"`: {
						{comment: "//nolint", sameLine: true},
						{comment: "// test 2", sameLine: false},
					},
				},
				alias: map[string]string{},
			},
			//
			want: `
	// import
	// sql
	"database/sql" //nolint:golint
	// fmt package
	"fmt"

	// test
	"github.com/remote/repo" //nolint

	// test 2
	"github.com/local/repo" //nolint
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := tt.pkg.fmt()
			want := strings.TrimPrefix(tt.want, "\n")
			require.Equal(t, want, string(got))
		})
	}
}

func TestProcessFile(t *testing.T) {
	t.Parallel()

	newBool := func(v bool) *bool { return &v }
	flagSet := &FlagSet{
		LocalFlag: "github.com/local/repo",
		DoWrite:   newBool(false),
		DoDiff:    newBool(false),
	}

	testNumbers := []int{1, 2}
	for _, testNumber := range testNumbers {
		testNumber := testNumber
		t.Run("", func(t *testing.T) {
			require := require.New(t)

			inFilepath := fmt.Sprintf("testdata/%d.in.go", testNumber)
			wantFilepath := fmt.Sprintf("testdata/%d.want.go", testNumber)

			wantFile, err := os.Open(wantFilepath)
			require.Nil(err)
			defer wantFile.Close()

			want, err := ioutil.ReadAll(wantFile)
			require.Nil(err)

			buf := bytes.NewBuffer(nil)
			err = ProcessFile(inFilepath, buf, flagSet)
			require.Nil(err)

			require.Equal(string(want), buf.String())
		})
	}
}
