package config

import (
	"fmt"
	"os"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/encoding/yaml"
	"github.com/geekxflood/common/logging"
)

// Provider defines the interface for accessing configuration
type Provider interface {
	GetString(key string, defaultValue string) (string, error)
	GetInt(key string, defaultValue int) (int, error)
	GetBool(key string, defaultValue bool) (bool, error)
	GetDuration(key string, defaultValue time.Duration) (time.Duration, error)
	GetMap(key string) (map[string]interface{}, error)
	GetStringSlice(key string) ([]string, error)
	Exists(key string) bool
	Validate() error
}

// Manager implements the Provider interface
type Manager struct {
	schema []byte
	config map[string]interface{}
	logger logging.Logger
}

// Options for creating a new Manager
type Options struct {
	SchemaContent []byte
	ConfigPath    string
	Logger        logging.Logger
}

// NewManager creates a new configuration manager
func NewManager(opts Options) (*Manager, error) {
	mgr := &Manager{
		schema: opts.SchemaContent,
		logger: opts.Logger,
	}

	// Load configuration file if provided
	if opts.ConfigPath != "" {
		data, err := os.ReadFile(opts.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Parse YAML
		if err := yaml.Unmarshal(data, &mgr.config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	} else {
		mgr.config = make(map[string]interface{})
	}

	return mgr, nil
}

// Validate validates the configuration against the CUE schema
func (m *Manager) Validate() error {
	ctx := cuecontext.New()

	// Compile the schema
	schemaValue := ctx.CompileBytes(m.schema)
	if schemaValue.Err() != nil {
		return fmt.Errorf("failed to compile schema: %w", schemaValue.Err())
	}

	// Convert config to CUE value
	configValue := ctx.Encode(m.config)
	if configValue.Err() != nil {
		return fmt.Errorf("failed to encode config: %w", configValue.Err())
	}

	// Unify schema and config
	unified := schemaValue.Unify(configValue)
	if unified.Err() != nil {
		return fmt.Errorf("config validation failed: %w", unified.Err())
	}

	// Validate the unified value
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// GetString retrieves a string value from the configuration
func (m *Manager) GetString(key string, defaultValue string) (string, error) {
	val, ok := m.getValue(key)
	if !ok {
		return defaultValue, nil
	}

	str, ok := val.(string)
	if !ok {
		return defaultValue, fmt.Errorf("value at %s is not a string", key)
	}

	return str, nil
}

// GetInt retrieves an integer value from the configuration
func (m *Manager) GetInt(key string, defaultValue int) (int, error) {
	val, ok := m.getValue(key)
	if !ok {
		return defaultValue, nil
	}

	// Handle both int and float64 (JSON numbers)
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return defaultValue, fmt.Errorf("value at %s is not an integer", key)
	}
}

// GetBool retrieves a boolean value from the configuration
func (m *Manager) GetBool(key string, defaultValue bool) (bool, error) {
	val, ok := m.getValue(key)
	if !ok {
		return defaultValue, nil
	}

	b, ok := val.(bool)
	if !ok {
		return defaultValue, fmt.Errorf("value at %s is not a boolean", key)
	}

	return b, nil
}

// GetDuration retrieves a duration value from the configuration
func (m *Manager) GetDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	val, ok := m.getValue(key)
	if !ok {
		return defaultValue, nil
	}

	str, ok := val.(string)
	if !ok {
		return defaultValue, fmt.Errorf("value at %s is not a string", key)
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return defaultValue, fmt.Errorf("failed to parse duration at %s: %w", key, err)
	}

	return duration, nil
}

// GetMap retrieves a map value from the configuration
func (m *Manager) GetMap(key string) (map[string]interface{}, error) {
	val, ok := m.getValue(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	mapVal, ok := val.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value at %s is not a map", key)
	}

	return mapVal, nil
}

// GetStringSlice retrieves a string slice from the configuration
func (m *Manager) GetStringSlice(key string) ([]string, error) {
	val, ok := m.getValue(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	slice, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value at %s is not a slice", key)
	}

	result := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("value at %s[%d] is not a string", key, i)
		}
		result[i] = str
	}

	return result, nil
}

// Exists checks if a key exists in the configuration
func (m *Manager) Exists(key string) bool {
	_, ok := m.getValue(key)
	return ok
}

// getValue retrieves a value from the nested configuration map
func (m *Manager) getValue(key string) (interface{}, bool) {
	return getNestedValue(m.config, key)
}
