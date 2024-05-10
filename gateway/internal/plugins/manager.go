package plugin

import (
	"net/http"
)

type Plugin interface {
	Apply(next http.Handler) http.Handler
}

type ConfigurablePlugin interface {
	Plugin
	Configure(map[string]interface{}) error
}

type Manager struct {
	Plugins map[string]ConfigurablePlugin
}

func NewManager() *Manager {
	return &Manager{
		Plugins: make(map[string]ConfigurablePlugin),
	}
}

func (pm *Manager) RegisterPlugin(name string, plugin ConfigurablePlugin) {
	pm.Plugins[name] = plugin
}

func (pm *Manager) ApplyPlugins(names []string, handler http.Handler) http.Handler {
	for _, name := range names {
		if plugin, ok := pm.Plugins[name]; ok {
			handler = plugin.Apply(handler)
		}
	}
	return handler
}
