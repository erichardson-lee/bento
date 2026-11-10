package tar

import (
	"archive/tar"
	"fmt"
	"io"

	"github.com/warpstreamlabs/bento/internal/filepath/ifs"
)

type TarFiles map[string][]byte

func ReadTarFromFile(path string) (TarFiles, error) {
	// Load from a tar file
	tarFile, err := ifs.OS().Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: failed to read plugin file: %w", path, err)
	}

	reader := tar.NewReader(tarFile)

	// Read all files into memory, within the map
	files := make(TarFiles)
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
