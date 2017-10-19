package module

import (
	"testing"

	"github.com/hashicorp/terraform/svchost"
)

func mustHostname(h string) svchost.Hostname {
	host, err := svchost.ForComparison(h)
	if err != nil {
		panic(err)
	}
	return host
}

func TestParseRegistrySource(t *testing.T) {
	for _, tc := range []struct {
		source      string
		host        svchost.Hostname
		id          string
		err         bool
		notRegistry bool
	}{
		{ // simple source id
			source: "namespace/id/provider",
			id:     "namespace/id/provider",
		},
		{ // source with hostname
			source: "registry.com/namespace/id/provider",
			host:   mustHostname("registry.com"),
			id:     "namespace/id/provider",
		},
		{ // source with hostname and port
			source: "registry.com:4443/namespace/id/provider",
			host:   mustHostname("registry.com:4443"),
			id:     "namespace/id/provider",
		},
		{ // too many parts
			source:      "registry.com/namespace/id/provider/extra",
			notRegistry: true,
		},
		{ // local path
			source:      "./local/file/path",
			notRegistry: true,
		},
		{ // local path with hostname
			source:      "./registry.com/namespace/id/provider",
			notRegistry: true,
		},
		{ // full URL
			source:      "https://example.com/foo/bar/baz",
			notRegistry: true,
		},
		{ // punycode host not allowed in source
			source: "xn--80akhbyknj4f.com/namespace/id/provider",
			err:    true,
		},
		{ // simple source id with subdir
			source: "namespace/id/provider//subdir",
			id:     "namespace/id/provider",
		},
		{ // source with hostname and subdir
			source: "registry.com/namespace/id/provider//subdir",
			host:   mustHostname("registry.com"),
			id:     "namespace/id/provider",
		},
		{ // source with hostname
			source: "registry.com/namespace/id/provider",
			host:   mustHostname("registry.com"),
			id:     "namespace/id/provider",
		},
		{ // we special case github
			source:      "github.com/namespace/id/provider",
			notRegistry: true,
		},
		{ // we special case github ssh
			source:      "git@github.com:namespace/id/provider",
			notRegistry: true,
		},
		{ // we special case bitbucket
			source:      "bitbucket.org/namespace/id/provider",
			notRegistry: true,
		},
	} {
		t.Run(tc.source, func(t *testing.T) {
			host, id, err := parseRegistrySource(tc.source)
			if tc.notRegistry {
				if err != ErrNotRegistry {
					t.Fatalf("%q should not be a registry source, got err %v", tc.source, err)
				}
				return
			}

			if tc.err {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if tc.host != "" {
				if host != tc.host {
					t.Fatalf("expected host %q, got %q", tc.host, host)
				}
			}

			if tc.id != id {
				t.Fatalf("expected id %q, got %q", tc.id, id)
			}
		})
	}
}
