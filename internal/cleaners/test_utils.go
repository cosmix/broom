package cleaners

import (
	"github.com/cosmix/broom/internal/utils"
)

// MockUtilsRunner implements utils.UtilsRunner for testing
type MockUtilsRunner struct {
	RunWithIndicatorFunc func(command, message string) error
	RunFdOrFindFunc      func(path, args, message string, sudo bool) error
	RunWithOutputFunc    func(command string) (string, error)
	CommandExistsFunc    func(command string) bool
	Commands             []string
}

func (m *MockUtilsRunner) RunWithIndicator(command, message string) error {
	m.Commands = append(m.Commands, command)
	return m.RunWithIndicatorFunc(command, message)
}

func (m *MockUtilsRunner) RunFdOrFind(path, args, message string, sudo bool) error {
	command := "fd/find " + path + " " + args
	m.Commands = append(m.Commands, command)
	return m.RunFdOrFindFunc(path, args, message, sudo)
}

func (m *MockUtilsRunner) RunWithOutput(command string) (string, error) {
	m.Commands = append(m.Commands, command)
	return m.RunWithOutputFunc(command)
}

func (m *MockUtilsRunner) CommandExists(command string) bool {
	return m.CommandExistsFunc(command)
}

func setupTest() *MockUtilsRunner {
	mock := &MockUtilsRunner{
		RunWithIndicatorFunc: func(command, message string) error { return nil },
		RunFdOrFindFunc:      func(path, args, message string, sudo bool) error { return nil },
		RunWithOutputFunc:    func(command string) (string, error) { return "", nil },
		CommandExistsFunc:    func(command string) bool { return true },
	}
	utils.SetUtilsRunner(mock)
	return mock
}
