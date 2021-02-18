package testdata

import (
	_ "embed" //nolint:golint
	// Import fmt package for `Println` function
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // import PostgreSQL driver

	// First package
	"github.com/local/repo/pkg1"
	// Second package
	"github.com/local/repo/pkg2"
)
