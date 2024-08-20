package llm

type ModelConfigStore struct {
	configs map[string]ModelConfig
}

func NewModelConfigStore() *ModelConfigStore {
	return &ModelConfigStore{
		configs: configs,
	}
}

func NewModelConfigStoreWithConfigs(configs map[string]ModelConfig) *ModelConfigStore {
	return &ModelConfigStore{
		configs: configs,
	}
}

func (m *ModelConfigStore) GetConfig(configName string) (ModelConfig, bool) {
	config, ok := m.configs[configName]
	return config, ok
}

func (m *ModelConfigStore) GetConfigNames() []string {
	var configNames []string
	for k := range m.configs {
		configNames = append(configNames, k)
	}
	return configNames
}

func (m *ModelConfigStore) AddConfig(configName string, config ModelConfig) {
	m.configs[configName] = config
}
