package models

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

func TestDashboardIndexesHaveStableColumnsAndPreserveParticipantUniqueness(t *testing.T) {
	participantSchema, err := schema.Parse(&Participant{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	require.Equal(t, []string{"membership_id", "event_id"}, indexColumns(t, participantSchema, "idx_membership_event"))
	require.Equal(t, []string{"event_id"}, indexColumns(t, participantSchema, "idx_participants_event_id"))
	require.True(t, indexByName(t, participantSchema, "idx_membership_event").Class == "UNIQUE")
	require.Empty(t, indexByName(t, participantSchema, "idx_participants_event_id").Class)

	eventSchema, err := schema.Parse(&Event{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	require.Equal(t, []string{"semester_id", "state", "start_date"}, indexColumns(t, eventSchema, "idx_events_semester_state_start_date"))
}

func indexColumns(t *testing.T, modelSchema *schema.Schema, name string) []string {
	t.Helper()
	index := indexByName(t, modelSchema, name)
	columns := make([]string, len(index.Fields))
	for i, field := range index.Fields {
		columns[i] = field.DBName
	}
	return columns
}

func indexByName(t *testing.T, modelSchema *schema.Schema, name string) *schema.Index {
	t.Helper()
	for _, index := range modelSchema.ParseIndexes() {
		if index.Name == name {
			return index
		}
	}
	t.Fatalf("index %q missing from %s schema", name, modelSchema.Table)
	return nil
}
