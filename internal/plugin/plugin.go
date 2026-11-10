package plugin

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/warpstreamlabs/bento/internal/bundle"
	"github.com/warpstreamlabs/bento/internal/filepath/ifs"
	"github.com/warpstreamlabs/bento/internal/plugin/bpl"
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
		manifestBytes, wasmSource, err := loadPlugin(path)
		if err != nil {
			return nil, fmt.Errorf("plugin %v: %w", path, err)
		}

		manifest, manifestLints, err := runtime.ReadManifestYAML(manifestBytes)
		if err != nil {
			return nil, fmt.Errorf("plugin %v: %w", path, err)
		}

		for _, l := range manifestLints {
			lints = append(lints, fmt.Sprintf("plugin %v: %v", path, l))
		}

		plugin, err := rt.Register(manifest, wasmSource)
		if err != nil {
			return nil, fmt.Errorf("plugin %v failed to register: %w", manifest.Name, err)
		}

		if err := plugin.RegisterWith(env); err != nil {
			return nil, fmt.Errorf("plugin %v: failed to register with environment: %w", manifest.Name, err)
		}
	}

	return lints, nil
}

func loadPlugin(path string) ([]byte, runtime.Source, error) {
	if filepath.Ext(path) == ".bpl" {
		files, err := bpl.ReadBplFromFile(path)
		if err != nil {
			return nil, nil, err
		}
		manifestBytes, ok := files["plugin.yaml"]
		if !ok {
			return nil, nil, errors.New("missing manifest \"plugin.yaml\" file")
		}
		return manifestBytes, runtime.BentoPluginArchiveSource(files), nil
	}

	manifestBytes, err := ifs.ReadFile(ifs.OS(), filepath.Join(path, "plugin.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return manifestBytes, runtime.DirSource(path), nil
}
