package control

import (
	"github.com/Codename-Uranium/tunnel/pkg/xerror"
	"go.uber.org/zap"
)

type ServiceController interface {
	Shutdown() error
	Running() bool
}

type ServiceMap struct {
	services map[string]ServiceController
	order    []string
}

func NewServiceMap() *ServiceMap {
	return &ServiceMap{
		services: make(map[string]ServiceController),
		order:    make([]string, 0),
	}
}

func (m *ServiceMap) RegisterService(name string, service ServiceController) {
	_, ok := m.services[name]
	if ok {
		zap.L().Fatal("Service is already registered", zap.String("name", name))
	}

	if service == nil {
		zap.L().Fatal("Service is nil", zap.String("name", name))
	}

	m.services[name] = service
	m.order = append(m.order, name)

	zap.L().Info("Registered service", zap.String("name", name))
}

func (m *ServiceMap) Service(name string) ServiceController {
	c, ok := m.services[name]
	if !ok {
		zap.L().Fatal("Service is not registered", zap.String("name", name))
		return nil
	}

	return c
}

func (m *ServiceMap) Shutdown() error {
	for idx := range m.order {
		ridx := len(m.order) - 1 - idx
		name := m.order[ridx]
		zap.L().Info("Shutting down service", zap.String("name", name))

		service := m.services[name]
		if service == nil {
			return xerror.EInternalError("service is nil", nil, zap.String("name", name))
		}

		err := service.Shutdown()
		if err != nil {
			return xerror.EInternalError("service is failed to shutdown", err, zap.String("name", name))
		}

		if service.Running() {
			return xerror.EInternalError("service is still running", nil, zap.String("name", name))
		}

		m.services[name] = nil
	}

	m.order = m.order[:0]
	for name := range m.services {
		delete(m.services, name)
	}

	if (len(m.order) != 0) || (len(m.services) != 0) {
		return xerror.EInvalidArgument("services data is not empty after shutdown", nil, zap.Any("order", m.order), zap.Any("services", m.services))
	}

	return nil
}
