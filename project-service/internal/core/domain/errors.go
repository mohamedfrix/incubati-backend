package domain

import "errors"

// Errors:
var (
	ErrInvalidDateFormat = errors.New("invalid date format please respect: YYYY/MM/DD")
	ErrInvalidInputData = errors.New("invalide input data")
	ErrNotAuthorized = errors.New("you are not authorized to perform this action")
	ErrProjectNotFound	 = errors.New("project not found")
)



// controls permissions:
type PermissionContext struct {
	IsAuthenticated bool 
	Owner bool
	CanEditProject bool
}


func (p *PermissionContext) CanCreateProject() bool {
	if !p.IsAuthenticated {
		return false
	}
	return true
}

func (p *PermissionContext) CanUpdateProject() bool {
	if !p.IsAuthenticated {
		return false
	}

	if !p.Owner && !p.CanEditProject {
		return false
	}

	return true
}


func (p *PermissionContext) CanDeleteProject() bool {
	if !p.IsAuthenticated {
		return false
	}

	if !p.Owner {
		return false
	}

	return true
}
