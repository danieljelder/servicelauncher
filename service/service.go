package service

import (
	"errors"
	"fmt"
	"github.com/jideji/servicelauncher/procs"
	"os"
	"os/exec"
	"syscall"
)

// Service represents a service that can be started.
type ServiceImpl struct {
	name      string
	Pattern   string
	Command   string
	Directory string
	process   *procs.Process
}

type Services map[string]Service

type Service interface {
	IsRunning() bool
	Name() string
	Pid() (int, error)
	Start() error
	Stop() error
}

func NewService(
	name string,
	Pattern string,
	Command string,
	Directory string) Service {

	return &ServiceImpl{
		name:      name,
		Pattern:   Pattern,
		Command:   Command,
		Directory: Directory,
	}
}

// Start runs the service using the service command.
// It redirects stdout+stderr to /tmp/<servicename>.log.
func (s *ServiceImpl) Start() error {
	logfile, err := os.Create(fmt.Sprintf("/tmp/%s.log", s.name))
	if err != nil {
		return err
	}

	cmd := exec.Command("bash", "-c", s.Command)
	cmd.Dir = s.Directory
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = logfile
	cmd.Stderr = logfile

	err = cmd.Start()
	if err != nil {
		return err
	}

	p, err := procs.FindByPid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	s.process = p

	return nil
}

// Pid returns the process id of the running service.
// Returns an error if process is not running.
func (s *ServiceImpl) Pid() (int, error) {
	p := s.getProcess()
	if p == nil {
		return -1, errors.New("No process found.")
	}
	return p.Pid, nil
}

func (s *ServiceImpl) Name() string {
	return s.name
}

// IsRunning returns true if process is running.
func (s *ServiceImpl) IsRunning() bool {
	return s.getProcess() != nil
}

// Stop kills the running process.
func (s *ServiceImpl) Stop() error {
	p := s.getProcess()
	if err := p.Kill(); err != nil {
		return err
	}
	s.process = nil
	return nil
}

func (s *ServiceImpl) getProcess() *procs.Process {
	if s.process == nil {
		pr, err := procs.FindByCommandLine(s.Pattern)
		if err != nil {
			panic(err)
		}
		s.process = pr
	}
	return s.process
}
