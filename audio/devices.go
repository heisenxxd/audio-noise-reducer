package audio

import (
	"fmt"
	"github.com/gen2brain/malgo"
)

type DeviceInfo = malgo.DeviceInfo

type Manager struct {
	ctx *malgo.AllocatedContext
}

func NewManager() (*Manager, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar contexto malgo: %w", err)
	}

	return &Manager{ctx: ctx}, nil
}

func (m *Manager) Close() error {
	if m.ctx != nil {
		if err := m.ctx.Uninit(); err != nil {
			return fmt.Errorf("erro ao desinicializar contexto: %w", err)
		}
		m.ctx.Free()
	}
	return nil
}

func (m *Manager) GetCaptureDevices() ([]DeviceInfo, error) {
	devices, err := m.ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar dispositivos de captura: %w", err)
	}
	return devices, nil
}

func (m *Manager) GetPlaybackDevices() ([]DeviceInfo, error) {
	devices, err := m.ctx.Devices(malgo.Playback)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar dispositivos de reprodução: %w", err)
	}
	return devices, nil
}

func (m *Manager) Context() malgo.Context {
	return m.ctx.Context
}
