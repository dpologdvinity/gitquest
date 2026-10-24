package scripts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstaller(t *testing.T) {
	for _, tc := range []struct {
		name, system, arch, version, failure string
		custom                               bool
	}{
		{name: "latest Linux", system: "Linux", arch: "x86_64"},
		{name: "pinned macOS", system: "Darwin", arch: "arm64", version: "v9.8.7", custom: true},
		{name: "bad checksum", system: "Linux", arch: "x86_64", version: "9.8.7", failure: "checksum mismatch"},
		{name: "missing checksum", system: "Linux", arch: "x86_64", failure: "missing checksum"},
		{name: "download failure", system: "Linux", arch: "x86_64", failure: "download failed"},
		{name: "invalid archive", system: "Linux", arch: "x86_64", failure: "archive extraction failed"},
		{name: "unsupported OS", system: "FreeBSD", arch: "x86_64", failure: "unsupported operating system"},
		{name: "unsupported architecture", system: "Linux", arch: "riscv64", failure: "unsupported architecture"},
		{name: "invalid version", system: "Linux", arch: "x86_64", version: "../../other", failure: "invalid version"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			fixtures := t.TempDir()
			bin := t.TempDir()
			var archive bytes.Buffer
			gz := gzip.NewWriter(&archive)
			tw := tar.NewWriter(gz)
			payload := []byte("#!/bin/sh\nprintf 'gitquest fixture\\n'\n")
			if err := tw.WriteHeader(&tar.Header{Name: "gitquest", Mode: 0755, Size: int64(len(payload))}); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(payload); err != nil {
				t.Fatal(err)
			}
			if err := tw.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
			data := archive.Bytes()
			if tc.name == "invalid archive" {
				data = []byte("not an archive")
			}
			write := func(path, content string, mode os.FileMode) {
				t.Helper()
				if err := os.WriteFile(path, []byte(content), mode); err != nil {
					t.Fatal(err)
				}
			}
			write(filepath.Join(fixtures, "archive"), string(data), 0600)
			sum := fmt.Sprintf("%x", sha256.Sum256(data))
			if tc.name == "bad checksum" {
				sum = strings.Repeat("0", 64)
			}
			system, arch := "linux", "amd64"
			if tc.system == "Darwin" {
				system = "darwin"
			}
			if tc.arch == "arm64" {
				arch = "arm64"
			}
			name := fmt.Sprintf("gitquest_9.8.7_%s_%s.tar.gz", system, arch)
			checksumName := name
			if tc.name == "missing checksum" {
				checksumName = "different.tar.gz"
			}
			write(filepath.Join(fixtures, "checksums"), sum+"  "+checksumName+"\n", 0600)
			write(filepath.Join(bin, "uname"), "#!/bin/sh\ncase \"$1\" in -s) printf '%s\\n' \"$TEST_SYSTEM\";; -m) printf '%s\\n' \"$TEST_ARCH\";; *) exit 1;; esac\n", 0755)
			write(filepath.Join(bin, "curl"), `#!/bin/sh
output=
url=
while [ "$#" -gt 0 ]; do
 case "$1" in
  --output|-o) output=$2; shift 2;;
  --write-out|-w) shift 2;;
  --proto|--proto-redir|--retry) shift 2;;
  https://*) url=$1; shift;;
  *) shift;;
 esac
done
base=https://github.com/dpologdvinity/gitquest/releases
case "$url" in
 "$base/latest") printf '%s/tag/v9.8.7' "$base";;
 "$base/download/v9.8.7/checksums.txt") cp "$TEST_FIXTURES/checksums" "$output";;
 "$base/download/v9.8.7/$TEST_ASSET")
  if [ "$TEST_FAILURE" = "download failure" ]; then exit 22; fi
  cp "$TEST_FIXTURES/archive" "$output";;
 *) printf 'unexpected URL: %s\n' "$url" >&2; exit 22;;
esac
`, 0755)
			dest := filepath.Join(home, ".local", "bin")
			if tc.custom {
				dest = filepath.Join(home, "custom bin")
			}
			if err := os.MkdirAll(dest, 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(dest, "gitquest")
			write(target, "existing binary", 0755)
			cmd := exec.Command("sh", "../install.sh")
			cmd.Env = append(os.Environ(), "HOME="+home, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "GITQUEST_VERSION="+tc.version, "GITQUEST_INSTALL_DIR=", "TEST_SYSTEM="+tc.system, "TEST_ARCH="+tc.arch, "TEST_FIXTURES="+fixtures, "TEST_ASSET="+name, "TEST_FAILURE="+tc.name)
			if tc.custom {
				cmd.Env = append(cmd.Env, "GITQUEST_INSTALL_DIR="+dest)
			}
			out, err := cmd.CombinedOutput()
			if tc.failure != "" {
				if err == nil || !strings.Contains(string(out), tc.failure) {
					t.Fatalf("got %v %s; want %s", err, out, tc.failure)
				}
				content, e := os.ReadFile(target)
				if e != nil || string(content) != "existing binary" {
					t.Fatal("failed install changed existing binary", e)
				}
				return
			}
			if err != nil {
				t.Fatalf("install: %v\n%s", err, out)
			}
			if info, e := os.Stat(target); e != nil || info.Mode().Perm() != 0755 {
				t.Fatalf("installed mode: %v %v", info, e)
			}
			result, e := exec.Command(target).Output()
			if e != nil || string(result) != "gitquest fixture\n" {
				t.Fatal(string(result), e)
			}
		})
	}
}
