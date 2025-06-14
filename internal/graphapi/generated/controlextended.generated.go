// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgenerated

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/graphapi/model"
)

// region    ************************** generated!.gotpl **************************

// endregion ************************** generated!.gotpl **************************

// region    ***************************** args.gotpl *****************************

// endregion ***************************** args.gotpl *****************************

// region    ************************** directives.gotpl **************************

// endregion ************************** directives.gotpl **************************

// region    **************************** field.gotpl *****************************

// endregion **************************** field.gotpl *****************************

// region    **************************** input.gotpl *****************************

func (ec *executionContext) unmarshalInputCloneControlInput(ctx context.Context, obj any) (model.CloneControlInput, error) {
	var it model.CloneControlInput
	asMap := map[string]any{}
	for k, v := range obj.(map[string]any) {
		asMap[k] = v
	}

	fieldsInOrder := [...]string{"controlIDs", "standardID", "ownerID", "programID"}
	for _, k := range fieldsInOrder {
		v, ok := asMap[k]
		if !ok {
			continue
		}
		switch k {
		case "controlIDs":
			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("controlIDs"))
			data, err := ec.unmarshalOID2ᚕstringᚄ(ctx, v)
			if err != nil {
				return it, err
			}
			it.ControlIDs = data
		case "standardID":
			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("standardID"))
			data, err := ec.unmarshalOID2ᚖstring(ctx, v)
			if err != nil {
				return it, err
			}
			it.StandardID = data
		case "ownerID":
			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("ownerID"))
			data, err := ec.unmarshalOID2ᚖstring(ctx, v)
			if err != nil {
				return it, err
			}
			it.OwnerID = data
		case "programID":
			ctx := graphql.WithPathContext(ctx, graphql.NewPathWithField("programID"))
			data, err := ec.unmarshalOID2ᚖstring(ctx, v)
			if err != nil {
				return it, err
			}
			it.ProgramID = data
		}
	}

	return it, nil
}

// endregion **************************** input.gotpl *****************************

// region    ************************** interface.gotpl ***************************

// endregion ************************** interface.gotpl ***************************

// region    **************************** object.gotpl ****************************

// endregion **************************** object.gotpl ****************************

// region    ***************************** type.gotpl *****************************

func (ec *executionContext) unmarshalOCloneControlInput2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCloneControlInput(ctx context.Context, v any) (*model.CloneControlInput, error) {
	if v == nil {
		return nil, nil
	}
	res, err := ec.unmarshalInputCloneControlInput(ctx, v)
	return &res, graphql.ErrorOnPath(ctx, err)
}

// endregion ***************************** type.gotpl *****************************
