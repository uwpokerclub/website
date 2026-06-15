package inmemory

import (
	"api/internal/models"
	"api/internal/store"
	"fmt"
	"sort"
	"sync"
)

type inMemoryStructureRepository struct {
  mu         sync.RWMutex
  structures map[int32]*models.Structure
  nextID     int32
}

var _ store.StructureRepository = (*inMemoryStructureRepository)(nil)

func newStructureRepository() *inMemoryStructureRepository {
  return &inMemoryStructureRepository{
    structures: make(map[int32]*models.Structure),
  }
}

func NewStructureRepository() store.StructureRepository {
  return newStructureRepository()
}

func (r *inMemoryStructureRepository) clone() *inMemoryStructureRepository {
  r.mu.RLock()
  defer r.mu.RUnlock()

  c := &inMemoryStructureRepository{
    structures: make(map[int32]*models.Structure, len(r.structures)),
    nextID:     r.nextID,
  }
  for id, s := range r.structures {
    sc := *s
    if s.Blinds != nil {
      sc.Blinds = make([]models.Blind, len(s.Blinds))
      copy(sc.Blinds, s.Blinds)
    }
    c.structures[id] = &sc
  }
  return c
}

func (r *inMemoryStructureRepository) Create(structure *models.Structure) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  if structure.ID == 0 {
    r.nextID++
    structure.ID = r.nextID
  } else if _, exists := r.structures[structure.ID]; exists {
    return fmt.Errorf("structure with ID %d already exists", structure.ID)
  }

  copy := *structure
  r.structures[structure.ID] = &copy

  return nil
}

func (r *inMemoryStructureRepository) FindByID(id int32) (models.Structure, error) {
  r.mu.RLock()
  defer r.mu.RUnlock()

  structure, exists := r.structures[id]
  if !exists {
    return models.Structure{}, store.ErrNotFound
}

  return *structure, nil
}

func (r *inMemoryStructureRepository) List(pagination *models.Pagination) ([]models.Structure, int64, error) {
  r.mu.RLock()
  defer r.mu.RUnlock()

  var structures []models.Structure
  for _, structure := range r.structures {
    structures = append(structures, *structure)
  }

  sort.Slice(structures, func(i, j int) bool {
    return structures[i].ID > structures[j].ID
  })
  
  total := int64(len(structures))
  
  offset := 0
  if pagination.Offset != nil && *pagination.Offset > 0 {
    offset = *pagination.Offset
  }

  if offset >= len(structures) {
    return []models.Structure{}, total, nil
  }

  structures = structures[offset:]

  if pagination.Limit != nil && *pagination.Limit > 0 && *pagination.Limit < len(structures) {
    structures = structures[:*pagination.Limit]
  }

  return structures, total, nil
}

func (r *inMemoryStructureRepository) Update(structure *models.Structure) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  existing, exists := r.structures[structure.ID]
  if !exists {
    return store.ErrNotFound
  }

  existing.Name = structure.Name

  return nil
}

func (r *inMemoryStructureRepository) ReplaceBlindsByStructureID(structureID int32, blinds []models.Blind) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  structure, exists := r.structures[structureID]
  if !exists {
    return store.ErrNotFound
  }

  structure.Blinds = blinds
  return nil
}

func (r *inMemoryStructureRepository) Delete(id int32) error {
  r.mu.Lock()
  defer r.mu.Unlock()

  if _, exists := r.structures[id]; !exists {
    return store.ErrNotFound
  }

  delete(r.structures, id)
  
  return nil
}
