package models

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestAccountActivationAssociationMatchesReleasedForeignKey(t *testing.T) {
	activationSchema, err := schema.Parse(&AccountActivation{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	require.Equal(t, "varchar", string(activationSchema.FieldsByDBName["username"].DataType))
	activationRelation := activationSchema.Relationships.Relations["Transition"]
	require.NotNil(t, activationRelation)
	activationConstraint := activationRelation.ParseConstraint()
	require.NotNil(t, activationConstraint)
	require.Equal(t, "account_activations_transition_id_fkey", activationConstraint.Name)
	require.Equal(t, "CASCADE", activationConstraint.OnDelete)

	loginSchema, err := schema.Parse(&Login{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	relation := loginSchema.Relationships.Relations["AccountActivations"]
	require.NotNil(t, relation)
	require.Equal(t, schema.HasMany, relation.Type)
	constraint := relation.ParseConstraint()
	require.NotNil(t, constraint)
	require.Equal(t, "fk_account_activations_username", constraint.Name)
	require.Equal(t, "account_activations", constraint.Schema.Table)
	require.Equal(t, "logins", constraint.ReferenceSchema.Table)
	require.Equal(t, "username", constraint.ForeignKeys[0].DBName)
	require.Equal(t, "username", constraint.References[0].DBName)
	require.Equal(t, "CASCADE", constraint.OnDelete)
	require.Equal(t, "NO ACTION", constraint.OnUpdate)

	require.NotContains(t, activationSchema.Relationships.Relations, "Login")
}

func TestOfficerTransitionUsernameColumnsMatchReleasedVarcharTypes(t *testing.T) {
	transitionSchema, err := schema.Parse(&OfficerTransition{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	for _, column := range []string{
		"initiated_by",
		"president_username",
		"vice_president_username",
		"secretary_username",
		"treasurer_username",
	} {
		require.Equal(t, "varchar", string(transitionSchema.FieldsByDBName[column].DataType), column)
	}
}
