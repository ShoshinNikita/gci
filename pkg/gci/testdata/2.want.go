package testdata

import (
	// Import embed to embed static files:
	//
	//   - templates
	//   - css files
	//   - js files
	//
	_ "embed" //nolint:golint
	// Import fmt package for `Println` function
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib" // import PostgreSQL driver

	// First package
	"github.com/local/repo/pkg1"
	// Second package
	"github.com/local/repo/pkg2"
)
