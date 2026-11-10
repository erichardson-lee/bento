package bpl

import (
	"archive/tar"
	"fmt"
	"io"

	"github.com/warpstreamlabs/bento/internal/filepath/ifs"
)

type BentoPluginArchive map[string][]byte

func ReadBplFromFile(path string) (BentoPluginArchive, error) {
	// Load from a tar file
	// Expects config.yaml to be the config, and then loads the rest into memort in a map[Path]Data format
	tarFile, err := ifs.OS().Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: failed to read plugin file: %w", path, err)
	}

	reader := tar.NewReader(tarFile)

	// Read all files into memory, within the map
	files := make(BentoPluginArchive)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("plugin %v: failed to read plugin file: %w", path, err)
		}

		fileData := make([]byte, header.Size)
		for {
			bytes, err := reader.Read(fileData)
			if err != nil && err != io.EOF {
				return nil, fmt.Errorf("plugin %v: failed to read plugin file: %w", path, err)
			}
			if bytes == 0 || err == io.EOF {
				break
			}
		}
		files[header.Name] = fileData
	}

	return files, nil
}
