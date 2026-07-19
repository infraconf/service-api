package perm

import (
	"slices"

	servicev1 "github.com/infraconf/service-api/api"
)

type TargetRessource struct {
	Target *servicev1.PermissionTarget
	Scope  *servicev1.PermissionScope
}

const WILDCARD_SYMBOL = ""
const CONTEXT_SYMBOL = "#"

func isMatchingTarget(grant *servicev1.PermissionTarget, res *servicev1.PermissionTarget) bool {
	if grant == nil || res == nil {
		return false
	}
	return (grant.Module == WILDCARD_SYMBOL || grant.Module == res.Module) &&
		(grant.ResourceType == WILDCARD_SYMBOL || grant.ResourceType == res.ResourceType) &&
		(grant.Action == WILDCARD_SYMBOL || grant.Action == res.Action)
}

func isMatchingScope(grant *servicev1.PermissionScope, res *servicev1.PermissionScope, cctx *servicev1.CallerContext) bool {
	if grant == nil || res == nil {
		return false
	}
	var org, group, owner, object bool

	if grant.OrganizationId == CONTEXT_SYMBOL {
		org = res.OrganizationId == cctx.OrganizationId
	} else {
		org = grant.OrganizationId == WILDCARD_SYMBOL || grant.OrganizationId == res.OrganizationId
	}

	if grant.GroupId == CONTEXT_SYMBOL {
		group = slices.Contains(cctx.GroupIds, res.GroupId)
	} else {
		group = grant.GroupId == WILDCARD_SYMBOL || grant.GroupId == res.GroupId
	}

	if grant.OwnerId == CONTEXT_SYMBOL {
		owner = res.OwnerId == cctx.UserId
	} else {
		owner = grant.OwnerId == WILDCARD_SYMBOL || grant.OwnerId == res.OwnerId
	}

	object = grant.ObjectId == WILDCARD_SYMBOL || grant.ObjectId == res.ObjectId

	return org && group && owner && object
}

func isMatching(grant *servicev1.PermissionGrant, res *TargetRessource, cctx *servicev1.CallerContext) bool {
	return isMatchingTarget(grant.Target, res.Target) && isMatchingScope(grant.Scope, res.Scope, cctx)
}

func HasPermission(grants []*servicev1.PermissionGrant, res *TargetRessource, cctx *servicev1.CallerContext) bool {
	allow := false

	for _, grant := range grants {
		if isMatching(grant, res, cctx) {
			switch grant.Effect {
			case servicev1.PermissionEffect_PERMISSION_EFFECT_ALLOW:
				allow = true
			case servicev1.PermissionEffect_PERMISSION_EFFECT_DENY:
				return false
			}
		}
	}

	return allow
}
