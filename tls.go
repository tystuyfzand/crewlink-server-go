package server

import (
	"errors"
	"os"
	"path"
)

// findCertificates will find certificates with common names in the specified path
func findCertificates(certificatePath string) (string, string, error) {
	if s, err := os.Stat(certificatePath); os.IsNotExist(err) || !s.IsDir() {
		return "", "", errors.New("certificate path does not exist or is not a directory")
	}

	validPaths := [][]string{
		// Let's Encrypt certificates
		{"fullchain.pem", "privkey.pem"},
		// Standard "server" naming
		{"server.crt", "server.key"},
	}

	var err error

	for _, pair := range validPaths {
		if _, err = os.Stat(path.Join(certificatePath, pair[0])); os.IsNotExist(err) {
			continue
		}

		if _, err = os.Stat(path.Join(certificatePath, pair[1])); os.IsNotExist(err) {
			continue
		}

		return pair[0], pair[1], nil
	}

	return "", "", errors.New("unable to find certificates in path")
}
