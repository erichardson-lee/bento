package plugin

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/warpstreamlabs/bento/internal/bundle"
	"github.com/warpstreamlabs/bento/internal/filepath/ifs"
	"github.com/warpstreamlabs/bento/internal/plugin/runtime"
)

var GlobalRuntime runtime.Runtime = &noopRuntime{}

type noopRuntime struct{}

var errNoPluginRuntime = errors.New("no plugin runtime enabled")

func (n *noopRuntime) Register(_ *runtime.Manifest, _ runtime.Source) (runtime.Plugin, error) {
	return nil, errNoPluginRuntime
}

func (n *noopRuntime) Close(_ context.Context) error {
	return nil
}

// Register registers a plugin with env using the global runtime.
func Register(env *bundle.Environment, manifest *runtime.Manifest, source runtime.Source) error {
	p, err := GlobalRuntime.Register(manifest, source)
	if err != nil {
		return fmt.Errorf("plugin %v failed to register: %w", manifest.Name, err)
	}

	if err := p.RegisterWith(env); err != nil {
		return fmt.Errorf("plugin %v: failed to register with environment: %w", manifest.Name, err)
	}

	return nil
}

func InitPlugins(ctx context.Context, env *bundle.Environment, rt runtime.Runtime, pluginPaths ...string) ([]string, error) {
	if len(pluginPaths) == 0 {
		return nil, nil
	}

	var lints []string
	for _, path := range pluginPaths {
		if filepath.Ext(path) != ".bpl" {
			newlints, err := InitFolderPlugin(path, ctx, env, rt)
			if err != nil {
				return nil, err
			}
			lints = append(lints, newlints...)
		} else {
			newLints, err := InitTarPlugin(path, ctx, env, rt)
			if err != nil {
				return nil, err
			}

			lints = append(lints, newLints...)
		}

	}

	return lints, nil
}

func InitTarPlugin(path string, ctx context.Context, env *bundle.Environment, rt runtime.Runtime) ([]string, error) {
	// Load from a tar file
	tarFile, err := ifs.OS().Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: failed to read plugin file: %w", path, err)
	}

	reader := tar.NewReader(tarFile)

	// Read all files into memory, within the map
	files := make(runtime.TarMapSource)
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

	// Kinda just same as initFolderPlugin from here
	// ************************************************************************************************
	manifestBytes, ok := files["plugin.yaml"]
	if !ok {
		return nil, fmt.Errorf("plugin %v: missing plugin.yaml", path)
	}

	// Converts manifest from bytes to in-memory object
	manifest, manifestLints, err := runtime.ReadManifestYAML(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: %w", path, err)
	}

	var lints []string
	// Add all linting errors to the list
	for _, l := range manifestLints {
		lints = append(lints, fmt.Sprintf("plugin %v: %v", path, l))
	}

	// Register the plugin with a manifest and a string directory to load from
	plugin, err := rt.Register(manifest, files)
	if err != nil {
		return nil, fmt.Errorf("plugin %v failed to register: %w", manifest.Name, err)
	}

	// Register with envrionment
	if err := plugin.RegisterWith(env); err != nil {
		return nil, fmt.Errorf("plugin %v: failed to register with environment: %w", manifest.Name, err)
	}

	return lints, nil
}

func InitFolderPlugin(path string, ctx context.Context, env *bundle.Environment, rt runtime.Runtime) ([]string, error) {

	// Load from a folder
	manifestPath := filepath.Join(path, "plugin.yaml")
	manifestBytes, err := ifs.ReadFile(ifs.OS(), manifestPath)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: failed to read manifest: %w", path, err)
	}

	// Kinda just same as initTarPlugin from here
	// ************************************************************************************************

	// Converts manifest from bytes to in-memory object
	manifest, manifestLints, err := runtime.ReadManifestYAML(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("plugin %v: %w", path, err)
	}

	var lints []string
	// Add all linting errors to the list
	for _, l := range manifestLints {
		lints = append(lints, fmt.Sprintf("plugin %v: %v", path, l))
	}

	// Register the plugin with a manifest and a string directory to load from
	plugin, err := rt.Register(manifest, runtime.DirSource(path))
	if err != nil {
		return nil, fmt.Errorf("plugin %v failed to register: %w", manifest.Name, err)
	}

	// Register with envrionment
	if err := plugin.RegisterWith(env); err != nil {
		return nil, fmt.Errorf("plugin %v: failed to register with environment: %w", manifest.Name, err)
	}

	return lints, nil
}
