package config

// FlowwError is the base error for all floww application errors.
type FlowwError struct {
	Msg   string
	Cause error
}

func (e *FlowwError) Error() string { return e.Msg }

// Unwrap returns the wrapped cause, enabling errors.Is/errors.As chaining.
func (e *FlowwError) Unwrap() error { return e.Cause }

// ConfigError represents configuration loading or validation errors.
type ConfigError struct{ FlowwError }

// ConfigLoadError represents configuration file loading errors.
type ConfigLoadError struct{ ConfigError }

// WorkflowNotFoundError represents workflow file not found errors.
type WorkflowNotFoundError struct{ ConfigError }

// WorkflowSchemaError represents workflow schema validation errors.
type WorkflowSchemaError struct{ ConfigError }

// WorkspaceError represents workspace management errors.
type WorkspaceError struct{ FlowwError }

// AppLaunchError represents application launch errors.
type AppLaunchError struct{ FlowwError }

// Sentinel errors for simple checks using errors.Is.
var (
	ErrConfigLoad       = &ConfigLoadError{ConfigError: ConfigError{FlowwError: FlowwError{Msg: "config load error"}}}
	ErrWorkflowNotFound = &WorkflowNotFoundError{ConfigError: ConfigError{FlowwError: FlowwError{Msg: "workflow not found"}}}
	ErrWorkflowSchema   = &WorkflowSchemaError{ConfigError: ConfigError{FlowwError: FlowwError{Msg: "workflow schema error"}}}
	ErrWorkspace        = &WorkspaceError{FlowwError: FlowwError{Msg: "workspace error"}}
	ErrAppLaunch        = &AppLaunchError{FlowwError: FlowwError{Msg: "app launch error"}}
)
