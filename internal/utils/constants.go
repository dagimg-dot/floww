package utils

// ValidWorkspaceBackends is the set of valid workspace backends for O(1) lookup.
var ValidWorkspaceBackends = map[string]bool{
	"auto":     true,
	"hyprland": true,
	"niri":     true,
	"ewmh":     true,
	"wmctrl":   true,
}

// GeneralConfig holds the general configuration section.
type GeneralConfig struct {
	ShowNotifications bool   `yaml:"show_notifications" json:"show_notifications" toml:"show_notifications"`
	WorkspaceBackend  string `yaml:"workspace_backend" json:"workspace_backend" toml:"workspace_backend"`
}

// TimingConfig holds the timing configuration section.
type TimingConfig struct {
	WorkspaceSwitchWait float64 `yaml:"workspace_switch_wait" json:"workspace_switch_wait" toml:"workspace_switch_wait"`
	AppLaunchWait       float64 `yaml:"app_launch_wait" json:"app_launch_wait" toml:"app_launch_wait"`
	RespectAppWait      bool    `yaml:"respect_app_wait" json:"respect_app_wait" toml:"respect_app_wait"`
}

// DefaultConfig holds the default configuration structure matching Python's DEFAULT_CONFIG.
type DefaultConfig struct {
	General GeneralConfig `yaml:"general" json:"general" toml:"general"`
	Timing  TimingConfig  `yaml:"timing" json:"timing" toml:"timing"`
}

// DefaultConfigValues are the default configuration values matching Python's DEFAULT_CONFIG.
var DefaultConfigValues = DefaultConfig{
	General: GeneralConfig{
		ShowNotifications: true,
		WorkspaceBackend:  "auto",
	},
	Timing: TimingConfig{
		WorkspaceSwitchWait: 3,
		AppLaunchWait:       1,
		RespectAppWait:      true,
	},
}

// SampleApp describes a single application entry in a sample workflow.
type SampleApp struct {
	Name string   `yaml:"name" json:"name" toml:"name"`
	Exec string   `yaml:"exec" json:"exec" toml:"exec"`
	Args []string `yaml:"args,omitempty" json:"args,omitempty" toml:"args,omitempty"`
}

// SampleWorkspace describes a single workspace entry in a sample workflow.
type SampleWorkspace struct {
	Target int         `yaml:"target" json:"target" toml:"target"`
	Apps   []SampleApp `yaml:"apps" json:"apps" toml:"apps"`
}

// SampleWorkflow is the struct type for sample workflow content.
type SampleWorkflow struct {
	Description string            `yaml:"description" json:"description" toml:"description"`
	Workspaces  []SampleWorkspace `yaml:"workspaces" json:"workspaces" toml:"workspaces"`
}

// SampleWorkflowContent is the pre-built sample workflow matching Python's SAMPLE_WORKFLOW_CONTENT.
var SampleWorkflowContent = SampleWorkflow{
	Description: "An example workflow.",
	Workspaces: []SampleWorkspace{
		{
			Target: 1,
			Apps: []SampleApp{
				{Name: "Terminal", Exec: "gnome-terminal"},
			},
		},
		{
			Target: 2,
			Apps: []SampleApp{
				{Name: "Browser", Exec: "firefox", Args: []string{"https://github.com/dagimg-dot/floww"}},
			},
		},
	},
}

// FileType represents a configuration file type as a string enum.
type FileType string

const (
	FileTypeYAML FileType = "yaml"
	FileTypeJSON FileType = "json"
	FileTypeTOML FileType = "toml"
)
