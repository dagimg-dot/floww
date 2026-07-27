package launcher

import (
	"os"
	"os/exec"
	"testing"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	l := New()
	assert.NotNil(t, l)
	assert.Nil(t, l.RunCommand)
}

func TestLaunchApp_Binary(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "echo",
		Exec: "echo",
		Type: "binary",
		Args: []string{"hello"},
	}
	ok, err := l.LaunchApp(app)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestLaunchApp_Flatpak(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "flatpak test",
		Exec: "org.mozilla.firefox",
		Type: "flatpak",
		Args: []string{"--new-window", "https://example.com"},
	}
	ok, err := l.LaunchApp(app)
	t.Logf("flatpak result: ok=%v, err=%v", ok, err)
}

func TestLaunchApp_Snap(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "snap test",
		Exec: "snap-app",
		Type: "snap",
		Args: []string{"--version"},
	}
	ok, err := l.LaunchApp(app)
	t.Logf("snap result: ok=%v, err=%v", ok, err)
}

func TestLaunchApp_DefaultType(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "echo",
		Exec: "echo",
		Args: []string{"test"},
	}
	ok, err := l.LaunchApp(app)
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestLaunchApp_EmptyExec(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "empty",
		Exec: "",
		Type: "binary",
	}
	ok, err := l.LaunchApp(app)
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestLaunchApp_FileNotFound(t *testing.T) {
	l := New()
	app := workflow.App{
		Name: "nonexistent",
		Exec: "this-command-definitely-does-not-exist-12345",
		Type: "binary",
	}
	ok, err := l.LaunchApp(app)
	assert.False(t, ok)
	require.Error(t, err)
	var appLaunchErr *config.AppLaunchError
	assert.ErrorAs(t, err, &appLaunchErr)
	assert.Contains(t, err.Error(), "application not found")
}

func TestLaunchApp_AbsolutePathNotFound(t *testing.T) {
	var appLaunchErr *config.AppLaunchError
	l := New()
	app := workflow.App{
		Name: "nonexistent",
		Exec: "/nonexistent/path/to/binary",
		Type: "binary",
	}
	ok, err := l.LaunchApp(app)
	assert.False(t, ok)
	require.Error(t, err)
	assert.ErrorAs(t, err, &appLaunchErr)
}

func TestLaunchProcess_Basic(t *testing.T) {
	l := New()
	ok, err := l.LaunchProcess([]string{"echo", "hello world"})
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestLaunchProcess_EmptyCmd(t *testing.T) {
	l := New()
	ok, err := l.LaunchProcess([]string{})
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestLaunchProcess_FileNotFound(t *testing.T) {
	var appLaunchErr *config.AppLaunchError
	l := New()
	ok, err := l.LaunchProcess([]string{"this-command-does-not-exist-99999"})
	assert.False(t, ok)
	require.Error(t, err)
	assert.ErrorAs(t, err, &appLaunchErr)
}

func TestLaunchProcess_AbsolutePathNotFound(t *testing.T) {
	var appLaunchErr *config.AppLaunchError
	l := New()
	ok, err := l.LaunchProcess([]string{"/nonexistent-binary-path"})
	assert.False(t, ok)
	require.Error(t, err)
	assert.ErrorAs(t, err, &appLaunchErr)
}

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		input    string
		expected string
	}{
		{"~", home},
		{"~/dir/file", home + "/dir/file"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~/", home},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandTilde(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestExpandTilde_NoTilde(t *testing.T) {
	result := expandTilde("/absolute/path")
	assert.Equal(t, "/absolute/path", result)
	result2 := expandTilde("relative/path")
	assert.Equal(t, "relative/path", result2)
}

func TestLaunchProcess_StdoutDevNull(t *testing.T) {
	l := New()
	ok, err := l.LaunchProcess([]string{"echo", "should-not-appear"})
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestRunCommandInjection(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	l := &AppLauncher{
		RunCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			cmd := exec.Command("echo")
			return cmd
		},
	}

	app := workflow.App{
		Name: "test",
		Exec: "myapp",
		Type: "binary",
		Args: []string{"--flag", "value"},
	}

	ok, err := l.LaunchApp(app)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "myapp", capturedName)
	assert.Equal(t, []string{"--flag", "value"}, capturedArgs)
}

func TestRunCommand_SnapOverride(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	l := &AppLauncher{
		RunCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			return exec.Command("echo")
		},
	}

	app := workflow.App{
		Name: "snap test",
		Exec: "snap-app",
		Type: "snap",
		Args: []string{"--version"},
	}

	ok, err := l.LaunchApp(app)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "snap-app", capturedName)
	assert.Equal(t, []string{"--version"}, capturedArgs)
}

func TestRunCommand_FlatpakOverride(t *testing.T) {
	var capturedName string
	var capturedArgs []string

	l := &AppLauncher{
		RunCommand: func(name string, arg ...string) *exec.Cmd {
			capturedName = name
			capturedArgs = arg
			return exec.Command("echo")
		},
	}

	app := workflow.App{
		Name: "flatpak test",
		Exec: "org.mozilla.firefox",
		Type: "flatpak",
		Args: []string{"--new-window", "https://example.com"},
	}

	ok, err := l.LaunchApp(app)
	assert.True(t, ok)
	assert.NoError(t, err)
	assert.Equal(t, "flatpak", capturedName)
	assert.Equal(t, []string{"run", "org.mozilla.firefox", "--new-window", "https://example.com"}, capturedArgs)
}

func TestLaunchApp_AllTypesWithOverride(t *testing.T) {
	tests := []struct {
		name        string
		app         workflow.App
		wantCmdName string
		wantCmdArgs []string
	}{
		{
			name: "binary",
			app: workflow.App{
				Name: "test", Exec: "/usr/bin/myapp", Type: "binary", Args: []string{"-a", "-b"},
			},
			wantCmdName: "/usr/bin/myapp",
			wantCmdArgs: []string{"-a", "-b"},
		},
		{
			name: "binary default type",
			app: workflow.App{
				Name: "test", Exec: "/usr/bin/myapp", Args: []string{"-a"},
			},
			wantCmdName: "/usr/bin/myapp",
			wantCmdArgs: []string{"-a"},
		},
		{
			name: "flatpak",
			app: workflow.App{
				Name: "test", Exec: "org.app.ID", Type: "flatpak", Args: []string{"--arg"},
			},
			wantCmdName: "flatpak",
			wantCmdArgs: []string{"run", "org.app.ID", "--arg"},
		},
		{
			name: "snap",
			app: workflow.App{
				Name: "test", Exec: "snap-app", Type: "snap", Args: []string{"--arg"},
			},
			wantCmdName: "snap-app",
			wantCmdArgs: []string{"--arg"},
		},
		{
			name: "binary no args",
			app: workflow.App{
				Name: "test", Exec: "/bin/true", Type: "binary",
			},
			wantCmdName: "/bin/true",
			wantCmdArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedName string
			var capturedArgs []string
			l := &AppLauncher{
				RunCommand: func(name string, arg ...string) *exec.Cmd {
					capturedName = name
					capturedArgs = arg
					cmd := exec.Command("echo")
					return cmd
				},
			}
			ok, err := l.LaunchApp(tt.app)
			assert.True(t, ok)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCmdName, capturedName)
			assert.Equal(t, tt.wantCmdArgs, capturedArgs)
		})
	}
}

func TestLaunchProcess_TildeExpansion(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	l := &AppLauncher{
		RunCommand: func(name string, arg ...string) *exec.Cmd {
			assert.Equal(t, home+"/bin/myapp", name)
			assert.Equal(t, home+"/data/file", arg[0])
			return exec.Command("echo")
		},
	}

	ok, err := l.LaunchProcess([]string{"~/bin/myapp", "~/data/file"})
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestLaunchProcess_OnlyTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	l := &AppLauncher{
		RunCommand: func(name string, arg ...string) *exec.Cmd {
			assert.Equal(t, home, name)
			return exec.Command("echo")
		},
	}

	ok, err := l.LaunchProcess([]string{"~"})
	assert.True(t, ok)
	assert.NoError(t, err)
}
