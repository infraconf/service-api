package perm

import (
	"errors"
	"fmt"
	"strings"

	servicev1 "github.com/infraconf/service-api/api"
)

var manage_actions = []string{"read", "list", "create", "update", "delete", "manage"}

func parsePermissionTarget(target string) ([]*servicev1.PermissionTarget, error) {
	parts := strings.Split(target, ":")

	if len(parts) != 3 {
		return nil, errors.New("invalid grant target format. Expected '<module>:<resourceType>:<action>'")
	}

	if parts[2] == "manage" {
		list := make([]*servicev1.PermissionTarget, len(manage_actions))
		for i := 0; i < len(manage_actions); i++ {
			list[i] = &servicev1.PermissionTarget{
				Namespace:    parts[0],
				ResourceType: parts[1],
				Action:       manage_actions[i],
			}
		}
		return list, nil
	} else {
		return []*servicev1.PermissionTarget{
			{
				Namespace:    parts[0],
				ResourceType: parts[1],
				Action:       parts[2],
			},
		}, nil
	}
}

func parsePermissionScope(scope string) (*servicev1.PermissionScope, error) {
	parts := strings.Split(scope, ":")

	if len(parts) != 4 {
		return nil, errors.New("invalid grant scope format. Expected '<orgId>:<groupId>:<ownerId>:<objectId>'")
	}

	return &servicev1.PermissionScope{
		TenantId: parts[0],
		GroupId:  parts[1],
		OwnerId:  parts[2],
		ObjectId: parts[3],
	}, nil
}

func ParsePermissionGrant(grant string) ([]*servicev1.PermissionGrant, error) {
	effect := servicev1.PermissionEffect_PERMISSION_EFFECT_ALLOW

	if strings.HasPrefix(grant, "!") {
		effect = servicev1.PermissionEffect_PERMISSION_EFFECT_DENY
		grant = grant[1:]
	}

	parts := strings.Split(strings.ToLower(grant), "@")
	if len(parts) != 2 {
		return nil, errors.New("invalid grant format. Expected '<module>:<resourceType>:<action>@<orgId>:<groupId>:<ownerId>:<objectId>'")
	}

	targets, err := parsePermissionTarget(parts[0])
	if err != nil {
		return nil, err
	}

	scope, err := parsePermissionScope(parts[1])
	if err != nil {
		return nil, err
	}

	list := make([]*servicev1.PermissionGrant, len(targets))
	for i := 0; i < len(targets); i++ {
		list[i] = &servicev1.PermissionGrant{
			Effect: effect,
			Target: targets[i],
			Scope:  scope,
		}
	}

	return list, nil
}

func ParsePermissionPatterns(patterns []string) ([]*servicev1.PermissionGrant, error) {
	grants := make([]*servicev1.PermissionGrant, 0, len(patterns))

	for i, pattern := range patterns {
		parsed, err := ParsePermissionGrant(pattern)
		if err != nil {
			return nil, fmt.Errorf("parse permission pattern %d %q: %w", i, pattern, err)
		}

		grants = append(grants, parsed...)
	}

	return grants, nil
}
