package modelparse

import (
	"os"
	"strings"

	"github.com/samber/lo"
)

type RoleInfo struct {
	ViewRoles    map[string][]string
	CrudRoles    map[string][]string
	InheritRoles map[string][]string
	CreateRoles  map[string][]string
}

const (
	crudAnnotation    = "# @crud:"
	viewAnnotation    = "# @view:"
	inheritAnnotation = "# @inherit:"
	createAnnotation  = "# @create:"
)

// ParseRoleAnnotations parses relevant annotations and role line from roles/roles.fga to determine which objects to add the roles to
func ParseRoleAnnotations(rolesFile string) (*RoleInfo, error) {
	data, err := os.ReadFile(rolesFile)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")

	crudMap := make(map[string][]string)
	viewMap := make(map[string][]string)
	inheritMap := make(map[string][]string)
	createMap := make(map[string][]string)

	var pendingCrud, pendingView, pendingInherit, pendingCreate []string

	isSeparator := func(c rune) bool {
		return c == ',' || c == ';' || c == ' '
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, crudAnnotation) {
			pendingCrud = strings.FieldsFunc(strings.TrimPrefix(line, crudAnnotation), isSeparator)
		}

		if strings.HasPrefix(line, viewAnnotation) {
			pendingView = strings.FieldsFunc(strings.TrimPrefix(line, viewAnnotation), isSeparator)
		}

		if strings.HasPrefix(line, inheritAnnotation) {
			pendingInherit = strings.FieldsFunc(strings.TrimPrefix(line, inheritAnnotation), isSeparator)
		}

		if strings.HasPrefix(line, createAnnotation) {
			pendingCreate = strings.FieldsFunc(strings.TrimPrefix(line, createAnnotation), isSeparator)
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
			}

			// reset pending annotations after processing a role definition
			pendingCrud = nil
			pendingView = nil
			pendingInherit = nil
			pendingCreate = nil
		}
	}

	return &RoleInfo{
		CrudRoles:    crudMap,
		ViewRoles:    viewMap,
		InheritRoles: inheritMap,
		CreateRoles:  createMap,
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
