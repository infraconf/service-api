package configs

type GatewayConfig struct {
	BackendRuntimes []BackendRuntimeRegistration `yaml:"backend_runtimes"`
}

type APISelector struct {
	OperationID string `yaml:"operation_id"`
}

type HTTPSelector struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
}

type OperationSelector struct {
	API  APISelector  `yaml:"api"`
	HTTP HTTPSelector `yaml:"http"`
}

type AuthBlock struct {
	RequiredPermissions []string `yaml:"required_permissions"`
	ForwardPermissions  []string `yaml:"forward_permissions"`
}

type BackendOperations struct {
	OperationID string            `yaml:"operation_id"`
	Selector    OperationSelector `yaml:"selector"`
	Auth        AuthBlock         `yaml:"auth"`
}

type BackendRuntimeRegistration struct {
	Name                   string              `yaml:"name"`
	InternalServiceID      string              `yaml:"internal_service_id"`
	InternalServiceAddress string              `yaml:"internal_service_addr"`
	Operations             []BackendOperations `yaml:"operations"`
}
