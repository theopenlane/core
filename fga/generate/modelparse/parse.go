package modelparse

import (
	"os"
	"strings"

	"github.com/samber/lo"
)

type RoleInfo struct {
	ViewRoles         map[string][]string
	CrudRoles         map[string][]string
	InheritRoles      map[string][]string
	CreateRoles       map[string][]string
	OrganizationRoles []OrganizationRole
}

type OrganizationRole struct {
	ID          string
	Name        string
	Description string
}

const (
	crudAnnotation    = "# @crud:"
	viewAnnotation    = "# @view:"
	inheritAnnotation = "# @inherit:"
	createAnnotation  = "# @create:"
	roleAnnotation    = "# @role:"
)

// ParseRoleAnnotations parses relevant annotations and role line from roles/roles.fga to determine which objects to add the roles to
func ParseRoleAnnotations(rolesFile string) (*RoleInfo, error) {
	data, err := os.ReadFile(rolesFile)
	if err != nil {
		return nil, err
	}

	return ParseRoleAnnotationsData(data)
}

func ParseRoleAnnotationsData(data []byte) (*RoleInfo, error) {
	lines := strings.Split(string(data), "\n")

	crudMap := make(map[string][]string)
	viewMap := make(map[string][]string)
	inheritMap := make(map[string][]string)
	createMap := make(map[string][]string)
	organizationRoles := []OrganizationRole{}

	var pendingCrud, pendingView, pendingInherit, pendingCreate []string
	var pendingRoleName, pendingRoleDescription string

	isSeparator := func(c rune) bool {
		return c == ',' || c == ';' || c == ' '
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if annotationValue, ok := strings.CutPrefix(line, crudAnnotation); ok {
			pendingCrud = strings.FieldsFunc(annotationValue, isSeparator)
		}

		if annotationValue, ok := strings.CutPrefix(line, viewAnnotation); ok {
			pendingView = strings.FieldsFunc(annotationValue, isSeparator)
		}

		if annotationValue, ok := strings.CutPrefix(line, inheritAnnotation); ok {
			pendingInherit = strings.FieldsFunc(annotationValue, isSeparator)
		}

		if annotationValue, ok := strings.CutPrefix(line, createAnnotation); ok {
			pendingCreate = strings.FieldsFunc(annotationValue, isSeparator)
		}

		// roles are defined as "# @role: Group Manager | Manage organization groups"
		if annotationValue, ok := strings.CutPrefix(line, roleAnnotation); ok {
			parts := strings.SplitN(strings.TrimSpace(annotationValue), "|", 2) //nolint:mnd
			pendingRoleName = strings.TrimSpace(parts[0])
			pendingRoleDescription = ""

			if len(parts) == 2 { //nolint:mnd
				pendingRoleDescription = strings.TrimSpace(parts[1])
			}
		}

		if strings.HasPrefix(line, "define ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				role := strings.TrimSuffix(parts[1], ":")

				for _, obj := range pendingCrud {
					crudMap[obj] = append(crudMap[obj], role)
				}

				for _, obj := range pendingView {
					viewMap[obj] = append(viewMap[obj], role)
				}

				for _, obj := range pendingCreate {
					createMap[obj] = append(createMap[obj], role)

				}

				inheritMap[role] = append(inheritMap[role], pendingInherit...)

				if pendingRoleName != "" {
					organizationRoles = append(organizationRoles, OrganizationRole{
						ID:          role,
						Name:        pendingRoleName,
						Description: pendingRoleDescription,
					})
				}
			}

			// reset pending annotations after processing a role definition
			pendingCrud = nil
			pendingView = nil
			pendingInherit = nil
			pendingCreate = nil
			pendingRoleName = ""
			pendingRoleDescription = ""
		}
	}

	return &RoleInfo{
		CrudRoles:         crudMap,
		ViewRoles:         viewMap,
		InheritRoles:      inheritMap,
		CreateRoles:       createMap,
		OrganizationRoles: organizationRoles,
	}, nil
}

// addInheritedRoles adds roles to objects based on the inherit annotations, so if role A inherits role B, and role B has access to edit a group, then role A should also have access to edit the group, even if it is not explicitly defined in the annotations for role A
func (r *RoleInfo) AddInheritedRoles() {
	for role, inherited := range r.InheritRoles {
		for _, in := range inherited {
			for obj := range r.CrudRoles {
				for _, rr := range r.CrudRoles[obj] {
					if rr == in {
						r.CrudRoles[obj] = append(r.CrudRoles[obj], role)
						break
					}
				}
			}

			for obj := range r.ViewRoles {
				for _, rr := range r.ViewRoles[obj] {
					if rr == in {
						r.ViewRoles[obj] = append(r.ViewRoles[obj], role)
						break
					}
				}
			}

			for obj := range r.CreateRoles {
				for _, rr := range r.CreateRoles[obj] {
					if rr == in {
						r.CreateRoles[obj] = append(r.CreateRoles[obj], role)
						break
					}
				}
			}
		}
	}

	// deduplicate roles for each object
	for obj := range r.CrudRoles {
		r.CrudRoles[obj] = lo.Uniq(r.CrudRoles[obj])
	}

	for obj := range r.ViewRoles {
		r.ViewRoles[obj] = lo.Uniq(r.ViewRoles[obj])
	}

	for obj := range r.CreateRoles {
		r.CreateRoles[obj] = lo.Uniq(r.CreateRoles[obj])
	}
}
