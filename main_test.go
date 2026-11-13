package main

import (
	"runtime/debug"
	"testing"
)

// TestModuleVersion pins how a build without release ldflags (for example
// `go install ...@v0.1.2`) gets its version from the embedded module info.
func TestModuleVersion(t *testing.T) {
	for _, c := range []struct {
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{&debug.BuildInfo{Main: debug.Module{Version: "v0.1.2"}}, true, "0.1.2"},
		{&debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true, ""},
		{&debug.BuildInfo{}, true, ""},
		{nil, false, ""},
	} {
		if got := moduleVersion(c.info, c.ok); got != c.want {
			t.Fatalf("moduleVersion(%+v, %v) = %q, want %q", c.info, c.ok, got, c.want)
		}
	}
}

// TestResolveVersion checks that a release version set with -ldflags always
// wins and that a local build with no module version stays "dev".
func TestResolveVersion(t *testing.T) {
	stamped := func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.1.2"}}, true
	}
	local := func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}
	for _, c := range []struct {
		version string
		read    func() (*debug.BuildInfo, bool)
		want    string
	}{
		{"0.1.3", stamped, "0.1.3"},
		{"dev", stamped, "0.1.2"},
		{"dev", local, "dev"},
	} {
		if got := resolveVersion(c.version, c.read); got != c.want {
			t.Fatalf("resolveVersion(%q) = %q, want %q", c.version, got, c.want)
		}
	}
}
