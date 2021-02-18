package testdata

import (
	// Import fmt package for `Println` function
	"fmt"
	_ "embed" //nolint:golint
	// Second package
	"github.com/local/repo/pkg2"
	// First package
	"github.com/local/repo/pkg1"
	_ "github.com/jackc/pgx/v4/stdlib" // import PostgreSQL driver
)
