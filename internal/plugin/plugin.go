package plugin

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/warpstreamlabs/bento/internal/bundle"
	"github.com/warpstreamlabs/bento/internal/filepath/ifs"
	"github.com/warpstreamlabs/bento/internal/filepath/tar"
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
		var files tar.TarFiles
		var manifestBytes []byte
		var source runtime.Source
		var err error
		var ok bool

		if filepath.Ext(path) == ".bpl" {
			files, err = tar.ReadTarFromFile(path)
			if err != nil {
				return nil, err
			}
			manifestBytes, ok = files["plugin.yaml"]
			if !ok {
				return nil, fmt.Errorf("plugin %v: missing manifest \"plugin.yaml\" file", path)
			}
			source = runtime.TarMapSource(files)
		} else {
			manifestPath := filepath.Join(path, "plugin.yaml")
			manifestBytes, err = ifs.ReadFile(ifs.OS(), manifestPath)
			if err != nil {
				return nil, fmt.Errorf("plugin %v: failed to read manifest: %w", path, err)
			}
			source = runtime.DirSource(path)
		}

		// Converts manifest from bytes to in-memory object
		manifest, manifestLints, err := runtime.ReadManifestYAML(manifestBytes)
		if err != nil {
			return nil, fmt.Errorf("plugin %v: %w", path, err)
		}

		// Add all linting errors to the list
		for _, l := range manifestLints {
			lints = append(lints, fmt.Sprintf("plugin %v: %v", path, l))
		}

		// Register the plugin with a manifest and a string directory to load from
		plugin, err := rt.Register(manifest, source)
		if err != nil {
			return nil, fmt.Errorf("plugin %v failed to register: %w", manifest.Name, err)
		}

		// Register with envrionment
		if err := plugin.RegisterWith(env); err != nil {
			return nil, fmt.Errorf("plugin %v: failed to register with environment: %w", manifest.Name, err)
		}
	}

	return lints, nil
}
