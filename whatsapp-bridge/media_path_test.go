package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// helper: create a temporary file inside dir and return its path.
func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestPathHasPrefix(t *testing.T) {
	sep := string(os.PathSeparator)
	cases := []struct {
		child, parent string
		want          bool
	}{
		{"/a/b", "/a/b", true},
		{"/a/b" + sep + "c", "/a/b", true},
		{"/a/bbb", "/a/b", false}, // sibling-prefix attack
		{"/a", "/a/b", false},     // shorter
		{"", "/a", false},
	}
	for _, tc := range cases {
		t.Run(tc.child+"|"+tc.parent, func(t *testing.T) {
			if got := pathHasPrefix(tc.child, tc.parent); got != tc.want {
				t.Fatalf("pathHasPrefix(%q,%q) = %v, want %v", tc.child, tc.parent, got, tc.want)
			}
		})
	}
}

func TestValidateMediaPathAcceptsFileInsideRoot(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("eval root: %v", err)
	}
	f := writeFile(t, root, "ok.txt", "hello")

	got, err := validateMediaPath(f, []string{resolvedRoot})
	if err != nil {
		t.Fatalf("expected accept, got error: %v", err)
	}
	resolvedFile, _ := filepath.EvalSymlinks(f)
	if got != resolvedFile {
		t.Fatalf("expected resolved path %q, got %q", resolvedFile, got)
	}
}

func TestValidateMediaPathRejectsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	other := t.TempDir()
	f := writeFile(t, other, "elsewhere.txt", "nope")

	if _, err := validateMediaPath(f, []string{resolvedRoot}); err == nil {
		t.Fatal("expected error for path outside root, got nil")
	}
}

func TestValidateMediaPathRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows; covered by Unix CI")
	}
	root := t.TempDir()
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	secrets := t.TempDir()
	target := writeFile(t, secrets, "id_rsa", "PRETEND-PRIVATE-KEY")

	// Drop a symlink inside the allowed root that points at a file outside.
	link := filepath.Join(root, "stolen")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	if _, err := validateMediaPath(link, []string{resolvedRoot}); err == nil {
		t.Fatal("expected symlink escape to be rejected, got nil error")
	} else if !strings.Contains(err.Error(), "outside the configured media roots") {
		t.Fatalf("expected confinement error, got: %v", err)
	}
}

func TestValidateMediaPathRejectsRelative(t *testing.T) {
	if _, err := validateMediaPath("not/absolute.txt", []string{"/tmp"}); err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestValidateMediaPathRejectsDirectory(t *testing.T) {
	root := t.TempDir()
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	subdir := filepath.Join(root, "sub")
	if err := os.Mkdir(subdir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := validateMediaPath(subdir, []string{resolvedRoot}); err == nil {
		t.Fatal("expected error for directory, got nil")
	}
}

func TestResolveMediaRootsRejectsRelativeEnv(t *testing.T) {
	t.Setenv("WHATSAPP_MEDIA_ROOTS", "relative/path")
	if _, err := resolveMediaRoots(); err == nil {
		t.Fatal("expected error for relative path in env, got nil")
	}
}

func TestResolveMediaRootsAcceptsEnvList(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	t.Setenv("WHATSAPP_MEDIA_ROOTS", a+string(os.PathListSeparator)+b)

	roots, err := resolveMediaRoots()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
}

func TestOutboundFileName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "windows absolute path, the reported leak",
			in:   `C:\Users\someone\.local\share\whatsapp-mcp\outbox\brief.zip`,
			want: "brief.zip",
		},
		{
			name: "windows path with spaces and parentheses",
			in:   `C:\Users\someone\outbox\2026-08-24 pitch (v19).mp4`,
			want: "2026-08-24 pitch (v19).mp4",
		},
		{
			name: "posix absolute path",
			in:   "/home/someone/.local/share/whatsapp-mcp/outbox/brief.zip",
			want: "brief.zip",
		},
		{
			name: "posix filename preserves colon",
			in:   "/home/someone/outbox/draft:v2.pdf",
			want: "draft:v2.pdf",
		},
		{
			name: "posix filename resembling a drive prefix is preserved",
			in:   "/home/someone/outbox/C:notes.pdf",
			want: "C:notes.pdf",
		},
		{
			name: "posix filename preserves surrounding spaces",
			in:   "/home/someone/outbox/ draft v2.pdf ",
			want: " draft v2.pdf ",
		},
		{
			name: "mixed separators, as a hand-written path often is",
			in:   `C:/Users/someone\outbox/brief.zip`,
			want: "brief.zip",
		},
		{
			name: "bare name, already safe",
			in:   "brief.zip",
			want: "brief.zip",
		},
		{
			name: "bare posix filename preserves colon",
			in:   "draft:v2.pdf",
			want: "draft:v2.pdf",
		},
		{
			name: "drive-relative path with no separator at all",
			in:   `C:brief.zip`,
			want: "brief.zip",
		},
		{
			name: "trailing separator yields a neutral fallback, never empty",
			in:   `C:\Users\someone\outbox\`,
			want: "file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := outboundFileName(tc.in)
			if got != tc.want {
				t.Fatalf("outboundFileName(%q) = %q, want %q", tc.in, got, tc.want)
			}
			// The real invariant: nothing that identifies the machine or the
			// user survives into the name, while valid filename characters stay intact.
			for _, leak := range []string{"Users", "home", "someone", `\`, "/"} {
				if strings.Contains(got, leak) {
					t.Fatalf("outboundFileName(%q) = %q, leaks %q", tc.in, got, leak)
				}
			}
		})
	}
}

func TestMediaDownloadPathsDefaultToStore(t *testing.T) {
	t.Setenv("WHATSAPP_MEDIA_DIR", "")
	ts := time.Date(2026, 11, 15, 7, 0, 0, 0, time.UTC)
	chatDir, localPath, _, err := mediaDownloadStorePaths("120363426758273498@g.us", "image", "ABC", ts)
	if err != nil {
		t.Fatalf("mediaDownloadStorePaths: %v", err)
	}
	storeRoot, _ := filepath.Abs("store")
	if want := filepath.Join(storeRoot, "120363426758273498@g.us"); chatDir != want {
		t.Fatalf("chatDir = %q, want %q", chatDir, want)
	}
	if !pathWithin(storeRoot, localPath) {
		t.Fatalf("localPath %q escapes %q", localPath, storeRoot)
	}
	if mediaDirMode() != 0o700 || mediaFileMode() != 0o600 {
		t.Fatalf("store/ media must stay owner-only, got %o/%o", mediaDirMode(), mediaFileMode())
	}
}

func TestMediaDownloadPathsHonourEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("WHATSAPP_MEDIA_DIR", "  "+root+"  ")
	ts := time.Date(2026, 11, 15, 7, 0, 0, 0, time.UTC)
	chatDir, localPath, filename, err := mediaDownloadStorePaths("447700900000:12@s.whatsapp.net", "document", "DEF", ts)
	if err != nil {
		t.Fatalf("mediaDownloadStorePaths: %v", err)
	}
	if want := filepath.Join(root, "447700900000_12@s.whatsapp.net"); chatDir != want {
		t.Fatalf("chatDir = %q, want %q", chatDir, want)
	}
	if want := filepath.Join(chatDir, filename); localPath != want {
		t.Fatalf("localPath = %q, want %q", localPath, want)
	}
	if mediaDirMode() != 0o755 || mediaFileMode() != 0o644 {
		t.Fatalf("shared media dir must be readable by other uids, got %o/%o", mediaDirMode(), mediaFileMode())
	}
}

func TestMediaDownloadPathsConfinedToMediaDir(t *testing.T) {
	t.Setenv("WHATSAPP_MEDIA_DIR", t.TempDir())
	ts := time.Date(2026, 11, 15, 7, 0, 0, 0, time.UTC)
	for _, id := range []string{"..", "../escape", "a/b"} {
		if _, _, _, err := mediaDownloadStorePaths("123@s.whatsapp.net", "image", id, ts); err == nil {
			t.Fatalf("message ID %q was accepted", id)
		}
	}
}
