package dockerctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"proyecto2-so1-201801391/internal/model"
)

type Manager struct {
	root        string
	carnet      string
	composeFile string
}

func New(root, carnet, composeFile string) *Manager {
	return &Manager{root: root, carnet: carnet, composeFile: composeFile}
}

func (m *Manager) StartInfrastructure(ctx context.Context) error {
	args := append(m.composeArgs(), "up", "-d", "valkey", "prometheus", "grafana")
	return m.run(ctx, args[0], args[1:]...)
}

func (m *Manager) BuildWorkloads(ctx context.Context) error {
	args := append(m.composeArgs(), "--profile", "build", "build")
	return m.run(ctx, args[0], args[1:]...)
}

func (m *Manager) LoadKernelModule(ctx context.Context) error {
	return m.run(ctx, "bash", filepath.Join(m.root, "scripts", "load_module.sh"))
}

func (m *Manager) InstallCron(ctx context.Context) error {
	return m.run(ctx, "bash", filepath.Join(m.root, "scripts", "install_cron.sh"))
}

func (m *Manager) RemoveCron(ctx context.Context) error {
	return m.run(ctx, "bash", filepath.Join(m.root, "scripts", "remove_cron.sh"))
}

func (m *Manager) EnsureBaseline(ctx context.Context) error {
	return m.run(ctx, "bash", filepath.Join(m.root, "scripts", "generate_containers.sh"), "--baseline")
}

func (m *Manager) ListManaged(ctx context.Context) ([]model.Container, error) {
	output, err := m.output(ctx, "docker", "ps", "-a", "--filter",
		"label=so1.project="+m.carnet, "--format", "{{.ID}}")
	if err != nil {
		return nil, err
	}
	ids := strings.Fields(string(output))
	containers := make([]model.Container, 0, len(ids))
	for _, id := range ids {
		container, err := m.inspect(ctx, id)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func (m *Manager) Stop(ctx context.Context, container model.Container) error {
	return m.run(ctx, "docker", "stop", "--time", "10", container.ID)
}

func (m *Manager) Remove(ctx context.Context, container model.Container) error {
	if container.ID == "" {
		return fmt.Errorf("cannot remove container without ID")
	}
	err := m.run(ctx, "docker", "rm", "-f", container.ID)
	if err == nil {
		return nil
	}
	// --rm containers can disappear concurrently with docker stop.
	// Verify absence through a successful Docker query; never hide daemon errors.
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		output, queryErr := m.output(waitCtx, "docker", "ps", "-a", "--no-trunc", "--filter", "id="+container.ID, "--format", "{{.ID}}")
		if queryErr != nil {
			return fmt.Errorf("verify container removal: %w", queryErr)
		}
		if len(strings.Fields(string(output))) == 0 {
			return nil
		}
		if !strings.Contains(err.Error(), "removal of container") || !strings.Contains(err.Error(), "already in progress") {
			return err
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait for container removal: %w", waitCtx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (m *Manager) Correlate(snapshot model.Snapshot, containers []model.Container) []model.ContainerMetric {
	metrics := make(map[string]*model.ContainerMetric, len(containers))
	for _, container := range containers {
		copy := container
		metrics[container.ID] = &model.ContainerMetric{Container: copy}
	}
	for _, process := range snapshot.Processes {
		cgroup, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", process.PID))
		container := matchContainer(string(cgroup), process.PID, containers)
		if container == nil {
			continue
		}
		metric := metrics[container.ID]
		metric.VSZKB += process.VSZKB
		metric.RSSKB += process.RSSKB
		metric.MemoryPercent += process.MemoryPercent
		metric.CPUPercent += process.CPUPercent
	}
	result := make([]model.ContainerMetric, 0, len(metrics))
	for _, metric := range metrics {
		result = append(result, *metric)
	}
	return result
}

func matchContainer(cgroup string, pid int, containers []model.Container) *model.Container {
	for index := range containers {
		container := &containers[index]
		if container.ID != "" && strings.Contains(cgroup, container.ID) {
			return container
		}
		if len(container.ID) >= 12 && strings.Contains(cgroup, container.ID[:12]) {
			return container
		}
	}
	for index := range containers {
		if containers[index].PID == pid {
			return &containers[index]
		}
	}
	return nil
}

func (m *Manager) inspect(ctx context.Context, id string) (model.Container, error) {
	type inspectResult struct {
		ID    string `json:"Id"`
		Name  string `json:"Name"`
		State struct {
			PID     int  `json:"Pid"`
			Running bool `json:"Running"`
		} `json:"State"`
		Config struct {
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
	}
	output, err := m.output(ctx, "docker", "inspect", id)
	if err != nil {
		return model.Container{}, err
	}
	var values []inspectResult
	if err := json.Unmarshal(output, &values); err != nil || len(values) != 1 {
		return model.Container{}, fmt.Errorf("decode docker inspect %s: %w", id, err)
	}
	value := values[0]
	return model.Container{
		ID: value.ID, Name: strings.TrimPrefix(value.Name, "/"), PID: value.State.PID,
		Running: value.State.Running, Labels: value.Config.Labels,
		Profile: value.Config.Labels["so1.profile"],
	}, nil
}

func (m *Manager) composeArgs() []string {
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return []string{"docker-compose", "--env-file", filepath.Join(m.root, ".env"), "-f", m.composeFile}
	}
	return []string{"docker", "compose", "--env-file", filepath.Join(m.root, ".env"), "-f", m.composeFile}
}

func (m *Manager) run(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = m.root
	command.Env = append(os.Environ(), "CARNET="+m.carnet, "PROJECT_ROOT="+m.root)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (m *Manager) output(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = m.root
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}
