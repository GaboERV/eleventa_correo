package extractor

import (
	"fmt"
	"io"
	"os"
)

// CopyFDB copies the Eleventa database file to a temporary location.
// This is critical: we NEVER operate on the original file.
// The original may be locked by Eleventa/AbarrotesPDV.
func CopyFDB(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source FDB %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot create destination FDB %s: %w", dst, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying FDB: %w", err)
	}

	return out.Sync()
}

// CleanupFDB removes the temporary copy of the database.
func CleanupFDB(path string) {
	os.Remove(path)
}
