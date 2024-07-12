package iostreams

import "os"

func hasAlternateScreenBuffer() bool {
	// on non-Windows, we just assume that alternate screen buffer is supported in most cases
	return os.Getenv("TERM") != "dumb"
}
