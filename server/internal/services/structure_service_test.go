package services

import (
	"api/internal/database"
	"api/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertBlindsEqual(t *testing.T, expected models.BlindJSON, actual models.Blind, msg string) {
	assert.Equal(t, expected.Small, actual.Small, msg)
	assert.Equal(t, expected.Big, actual.Big, msg)
	assert.Equal(t, expected.Ante, actual.Ante, msg)
	assert.Equal(t, expected.Time, actual.Time, msg)
}

func TestStructureService(t *testing.T) {
	t.Setenv("ENVIRONMENT", "TEST")

	db, err := database.OpenTestConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer sqlDB.Close()

	wipeDB := func() {
		err := database.WipeDB(db)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	structureService := NewStructureService(db)
	t.Run("CreateStructure", func(t *testing.T) {
		t.Cleanup(wipeDB)

		req := models.CreateStructureRequest{
			Name: "World Series of Poker",
			Blinds: []models.BlindJSON{
				{
					Small: 10,
					Big:   20,
					Ante:  0,
					Time:  20,
				},
				{
					Small: 20,
					Big:   40,
					Ante:  40,
					Time:  15,
				},
				{
					Small: 40,
					Big:   80,
					Ante:  80,
					Time:  10,
				},
			},
		}

		structure, err := structureService.CreateStructure(&req)
		assert.NoError(t, err, "CreateStructure should not error")
		assert.Equal(t, req.Name, structure.Name, "Structure name should be the same")
		assert.Len(t, structure.Blinds, len(req.Blinds), "Should have the same number of blind levels")
		assertBlindsEqual(t, req.Blinds[0], structure.Blinds[0], "Blind level 1 should be the same")
		assertBlindsEqual(t, req.Blinds[1], structure.Blinds[1], "Blind level 2 should be the same")
		assertBlindsEqual(t, req.Blinds[2], structure.Blinds[2], "Blind level 3 should be the same")
	})
	t.Run("ListStructures", func(t *testing.T) {
		t.Cleanup(wipeDB)

		testStructureA := models.Structure{
			Name: "Test Structure A",
		}
		res := db.Create(&testStructureA)
		assert.NoError(t, res.Error)
		testStructureB := models.Structure{
			Name: "Test Structure B",
		}
		res = db.Create(&testStructureB)
		assert.NoError(t, res.Error)
		testStructureC := models.Structure{
			Name: "Test Structure C",
		}
		res = db.Create(&testStructureC)
		assert.NoError(t, res.Error)

		structures, err := structureService.ListStructures()
		assert.NoError(t, err, "Should not return an error")
		assert.Len(t, structures, 3, "Should return 3 structures")
		assert.Equal(t, structures, []models.Structure{testStructureC, testStructureB, testStructureA})
	})
	t.Run("GetStructure", func(t *testing.T) {
		t.Cleanup(wipeDB)

		testStructure := models.Structure{
			Name: "Test Structure A",
		}
		res := db.Create(&testStructure)
		assert.NoError(t, res.Error)
		fakeStructure := models.Structure{
			Name: "Test Structure B",
		}
		res = db.Create(&fakeStructure)
		assert.NoError(t, res.Error)

		testBlinds := [3]models.Blind{
			{
				Small:       10,
				Big:         20,
				Ante:        0,
				Time:        20,
				Index:       0,
				StructureId: testStructure.ID,
			},
			{
				Small:       20,
				Big:         40,
				Ante:        40,
				Time:        15,
				Index:       1,
				StructureId: testStructure.ID,
			},
			{
				Small:       40,
				Big:         80,
				Ante:        80,
				Time:        10,
				Index:       2,
				StructureId: testStructure.ID,
			},
		}
		res = db.Create(&testBlinds)
		assert.NoError(t, res.Error)

		structure, err := structureService.GetStructure(testStructure.ID)
		assert.NoError(t, err)
		assert.Equal(t, testStructure.ID, structure.ID)
		assert.Equal(t, testStructure.Name, structure.Name)
		assert.Len(t, structure.Blinds, 3)
		assert.EqualValues(t, testBlinds[0], structure.Blinds[0])
		assert.EqualValues(t, testBlinds[1], structure.Blinds[1])
		assert.EqualValues(t, testBlinds[2], structure.Blinds[2])
	})
	t.Run("UpdateStructure", func(t *testing.T) {
		t.Cleanup(wipeDB)

		testStructure := models.Structure{
			Name: "Test Structure A",
		}
		res := db.Create(&testStructure)
		assert.NoError(t, res.Error)
		fakeStructure := models.Structure{
			Name: "Test Structure B",
		}
		res = db.Create(&fakeStructure)
		assert.NoError(t, res.Error)

		testBlinds := [3]models.Blind{
			{
				Small:       10,
				Big:         20,
				Ante:        0,
				Time:        20,
				Index:       0,
				StructureId: testStructure.ID,
			},
			{
				Small:       20,
				Big:         40,
				Ante:        40,
				Time:        15,
				Index:       1,
				StructureId: testStructure.ID,
			},
			{
				Small:       40,
				Big:         80,
				Ante:        80,
				Time:        10,
				Index:       2,
				StructureId: testStructure.ID,
			},
		}
		res = db.Create(&testBlinds)
		assert.NoError(t, res.Error)

		req := models.UpdateStructureRequest{
			ID:   testStructure.ID,
			Name: "New name",
			Blinds: []models.BlindJSON{
				{
					Small: 15,
					Big:   30,
					Ante:  0,
					Time:  15,
				},
				{
					Small: 25,
					Big:   50,
					Ante:  10,
					Time:  10,
				},
			},
		}

		structure, err := structureService.UpdateStructure(&req)
		assert.NoError(t, err)
		assert.Equal(t, testStructure.ID, structure.ID)
		assert.Equal(t, req.Name, structure.Name)
		assert.Len(t, structure.Blinds, 2)
		assertBlindsEqual(t, req.Blinds[0], structure.Blinds[0], "")
		assertBlindsEqual(t, req.Blinds[1], structure.Blinds[1], "")
		// Check that old blinds got deleted
		var blinds []models.Blind
		res = db.Where("structure_id = ?", testStructure.ID).Find(&blinds)
		assert.NoError(t, res.Error)
		assert.Len(t, blinds, 2)
	})
}
