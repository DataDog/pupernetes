package version

import "fmt"

var (
	// Package is filled at linking time
	Package = "github.com/DataDog/pupernetes"

	// Version holds the complete version number. Filled in at linking time.
	Version = "0.0.0+unknown"

	// Revision is filled with the VCS (e.g. git) revision being used to build
	// the program at linking time.
	Revision = ""
)

// DisplayVersion print to stdout the package/version/revision
func DisplayVersion() {
	fmt.Printf(`package: %s
version: %s
revision: %s
`, Package, Version, Revision)
}
