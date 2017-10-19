package module

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/svchost"
)

var (
	// these prefixes can't be registry IDs
	filterParts = []string{
		"https?:",         // http and https
		"[.]{0,2}/",       // local file paths
		"[A-Za-z0-9]+::",  // forced getter protocols
		"github.com/",     // github getter
		"git@github.com:", // github ssh getter
		"bitbucket.org/",  // bitbucket getter
	}

	skipRegistry = regexp.MustCompile(
		fmt.Sprintf("^(%s)", strings.Join(filterParts, "|")),
	).MatchString
)

// ErrNotRegistry is returned by parseRegistrySource for any string that can't
// be a registry ID.
var ErrNotRegistry = errors.New("not a registry source")

type moduleSource struct {
	host   svchost.Hostname
	module string
	subdir string
}

// ParseSource checks if a source string looks like a registry module source,
// and does some basic validation. It it succeeds, it returns a normalzed
// hostname and a module ID.
func parseRegistrySource(src string) (svchost.Hostname, string, error) {
	if skipRegistry(src) {
		return "", "", ErrNotRegistry
	}

	// Trim off any subdir.
	// This is done by the module loader for local storage too, so we don't
	// normally see these right now.
	src, _ = getter.SourceDirSubdir(src)

	// a registry source will have either 3 or 4 parts
	parts := strings.Split(src, "/")
	var host, id string
	switch len(parts) {
	case 3:
		// this is a short ID from the default registry
		id = src
	case 4:
		host = parts[0]
		id = strings.Join(parts[1:], "/")
	default:
		// We may be able to enforce this even more strictly and raise a
		// descriptive error.
		return "", "", ErrNotRegistry
	}

	if host == "" {
		return "", id, nil
	}

	hostname, err := svchost.ForComparison(host)
	if err != nil {
		return "", "", err
	}
	return hostname, id, nil
}
