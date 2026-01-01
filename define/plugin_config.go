package define

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// PluginConfig matches the host plugin config format.
// PathValue is runtime-only and is not part of the on-disk JSON.
type PluginConfig struct {
	Name        string                 `json:"名称"`
	Description string                 `json:"描述"`
	Author      string                 `json:"作者,omitempty"`
	Source      string                 `json:"来源"`
	Disable     bool                   `json:"是否禁用"`
	Version     string                 `json:"版本,omitempty"`
	Config      map[string]interface{} `json:"配置"`

	PathValue string `json:"-"`
}

func (pc *PluginConfig) Path() string {
	if pc == nil {
		return ""
	}
	return pc.PathValue
}

// LoadPluginConfig loads a plugin config from file and fills PathValue with an absolute path.
func LoadPluginConfig(filePath string) (*PluginConfig, error) {
	var errs []error

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		errs = append(errs, fmt.Errorf("get absolute path failed: %w", err))
	}

	var data []byte
	if len(errs) == 0 {
		data, err = os.ReadFile(absPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("read file failed: %w", err))
		}
	}

	var cfg PluginConfig
	if len(errs) == 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			errs = append(errs, fmt.Errorf("unmarshal json failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	cfg.PathValue = absPath
	return &cfg, nil
}
