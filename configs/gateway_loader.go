package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/infraconf/service-api/perm"
	"gopkg.in/yaml.v3"
)

func LoadGatewayConfigFile(path string) (*GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read gateway config file: %w", err)
	}

	cfg, err := LoadGatewayConfig(data)
	if err != nil {
		return nil, fmt.Errorf("load gateway config file: %w", err)
	}

	return cfg, nil
}

func LoadGatewayConfig(data []byte) (*GatewayConfig, error) {
	var cfg GatewayConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse gateway config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg GatewayConfig) Validate() error {
	operationIDs := map[string]struct{}{}
	apiSelectors := map[string]struct{}{}
	httpSelectors := map[string]struct{}{}

	for runtimeIndex, runtime := range cfg.BackendRuntimes {
		runtimeRef := fmt.Sprintf("backend_runtimes[%d]", runtimeIndex)

		if strings.TrimSpace(runtime.Name) == "" {
			return fmt.Errorf("%s.name is required", runtimeRef)
		}

		if strings.TrimSpace(runtime.InternalServiceID) == "" {
			return fmt.Errorf("%s.internal_service_id is required", runtimeRef)
		}

		if len(runtime.Operations) == 0 {
			return fmt.Errorf("%s.operations must not be empty", runtimeRef)
		}

		for operationIndex, operation := range runtime.Operations {
			operationRef := fmt.Sprintf("%s.operations[%d]", runtimeRef, operationIndex)
			if err := validateBackendOperation(operationRef, operation, operationIDs, apiSelectors, httpSelectors); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateBackendOperation(operationRef string, operation BackendOperations, operationIDs, apiSelectors, httpSelectors map[string]struct{}) error {
	operationID := strings.TrimSpace(operation.OperationID)
	if operationID == "" {
		return fmt.Errorf("%s.operation_id is required", operationRef)
	}

	if _, ok := operationIDs[operationID]; ok {
		return fmt.Errorf("%s.operation_id %q is already registered", operationRef, operationID)
	}
	operationIDs[operationID] = struct{}{}

	hasAPISelector := strings.TrimSpace(operation.Selector.API.OperationID) != ""
	hasHTTPSelector := strings.TrimSpace(operation.Selector.HTTP.Method) != "" || strings.TrimSpace(operation.Selector.HTTP.Path) != ""

	if !hasAPISelector && !hasHTTPSelector {
		return fmt.Errorf("%s.selector must define api.operation_id or http method and path", operationRef)
	}

	if hasAPISelector {
		apiOperationID := strings.TrimSpace(operation.Selector.API.OperationID)
		if apiOperationID != operationID {
			return fmt.Errorf("%s.selector.api.operation_id %q must match operation_id %q", operationRef, apiOperationID, operationID)
		}

		if _, ok := apiSelectors[apiOperationID]; ok {
			return fmt.Errorf("%s.selector.api.operation_id %q is already registered", operationRef, apiOperationID)
		}
		apiSelectors[apiOperationID] = struct{}{}
	}

	if hasHTTPSelector {
		method := strings.ToUpper(strings.TrimSpace(operation.Selector.HTTP.Method))
		path := strings.TrimSpace(operation.Selector.HTTP.Path)

		if method == "" || path == "" {
			return fmt.Errorf("%s.selector.http must define method and path together", operationRef)
		}

		selectorKey := method + " " + path
		if _, ok := httpSelectors[selectorKey]; ok {
			return fmt.Errorf("%s.selector.http %q is already registered", operationRef, selectorKey)
		}
		httpSelectors[selectorKey] = struct{}{}
	}

	if len(operation.Auth.RequiredPermissions) == 0 {
		return fmt.Errorf("%s.auth.required_permissions must not be empty", operationRef)
	}

	if _, err := perm.ParsePermissionPatterns(operation.Auth.RequiredPermissions); err != nil {
		return fmt.Errorf("%s.auth.required_permissions: %w", operationRef, err)
	}

	if _, err := perm.ParsePermissionPatterns(operation.Auth.ForwardPermissions); err != nil {
		return fmt.Errorf("%s.auth.forward_permissions: %w", operationRef, err)
	}

	return nil
}
