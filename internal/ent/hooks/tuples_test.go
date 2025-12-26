package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
)

func TestGetTupleKeyFromRole(t *testing.T) {
	testCases := []struct {
		name        string
		subID       string
		subType     string
		objID       string
		objType     string
		role        enums.Role
		expectedRes fgax.TupleKey
		expectedErr error
	}{
		{
			name:    "happy path",
			subID:   "01HM7RYYECMKN3FJWSAZVVQE4A",
			subType: "organization",
			objID:   "01HM7RVM7G2AVBQBTJA2TWCHHG",
			objType: "group",
			role:    fgax.ParentRelation,
			expectedRes: fgax.TupleKey{
				Subject: fgax.Entity{
					Kind:       "organization",
					Identifier: "01HM7RYYECMKN3FJWSAZVVQE4A",
				},
				Relation: "parent",
				Object: fgax.Entity{
					Kind:       "group",
					Identifier: "01HM7RVM7G2AVBQBTJA2TWCHHG",
				},
			},
		},
		{
			name:        "invalid role",
			subID:       "01HM7RYYECMKN3FJWSAZVVQE4A",
			subType:     "organization",
			objID:       "01HM7RVM7G2AVBQBTJA2TWCHHG",
			objType:     "group",
			role:        "baller",
			expectedRes: fgax.TupleKey{},
			expectedErr: ErrUnsupportedFGARole,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			req := fgax.TupleRequest{
				SubjectID:   tc.subID,
				SubjectType: tc.subType,
				ObjectID:    tc.objID,
				ObjectType:  tc.objType,
			}

			res, err := getTupleKeyFromRole(req, tc.role)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes, res)
		})
	}
}

func TestRoleToRelation(t *testing.T) {
	testCases := []struct {
		name        string
		roleInput   enums.Role
		expectedRes string
		expectedErr error
	}{
		{
			name:        "happy path, owner",
			roleInput:   enums.RoleOwner,
			expectedRes: "owner",
		},
		{
			name:        "happy path, admin",
			roleInput:   enums.RoleAdmin,
			expectedRes: "admin",
		},
		{
			name:        "happy path, member",
			roleInput:   enums.RoleMember,
			expectedRes: "member",
		},
		{
			name:        "happy path, parent",
			roleInput:   fgax.ParentRelation,
			expectedRes: "parent",
		},
		{
			name:        "invalid role",
			roleInput:   "baller",
			expectedRes: "",
			expectedErr: ErrUnsupportedFGARole,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			res, err := roleToRelation(tc.roleInput)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes, res)
		})
	}
}
