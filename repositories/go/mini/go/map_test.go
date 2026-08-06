package mini

import (
	"testing"
	"unsafe"
)

func TestNewMapStackPath(t *testing.T) {
	m := NewMap(0, false)
	if m == nil {
		t.Fatal("NewMap returned nil")
	}
	if m.seed == 0 {
		t.Fatal("NewMap did not initialize seed")
	}
	if m.dirPtr == nil {
		t.Fatal("non-escaping small map should initialize group")
	}
	if m.dirLen != 0 {
		t.Fatalf("small map dirLen = %d, want 0", m.dirLen)
	}
}

func TestNewMapHeapPath(t *testing.T) {
	m := NewMap(0, true)
	if m.seed == 0 {
		t.Fatal("NewMap did not initialize seed")
	}
	if m.dirPtr != nil {
		t.Fatal("small map should defer group allocation")
	}
}

func TestNewMapWithHint(t *testing.T) {
	m := NewMap(100, true)
	if m.dirPtr == nil {
		t.Fatal("NewMap did not allocate directory")
	}
	if m.dirLen != 1 {
		t.Fatalf("NewMap dirLen = %d, want 1", m.dirLen)
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	if directory[0] == nil {
		t.Fatal("directory[0] is nil")
	}
	if directory[0].capacity < 100 {
		t.Fatalf("table capacity = %d, want at least 100", directory[0].capacity)
	}
	if directory[0].groups.data == nil {
		t.Fatal("table groups are not initialized")
	}
}

func TestMapAccessIntSmallMap(t *testing.T) {
	m := NewMap(0, true)
	MapInsertInt(m, 1, 10)
	MapInsertInt(m, 2, 20)

	value, ok := MapAccessInt(m, 2)
	if !ok || value != 20 {
		t.Fatalf("MapAccessInt(2) = (%d, %v), want (20, true)", value, ok)
	}

	if _, ok := MapAccessInt(m, 3); ok {
		t.Fatal("MapAccessInt(3) found an absent key")
	}
}

func TestMapInsertIntOverwritesValue(t *testing.T) {
	m := NewMap(0, true)
	MapInsertInt(m, 1, 10)
	MapInsertInt(m, 1, 20)

	value, ok := MapAccessInt(m, 1)
	if !ok || value != 20 {
		t.Fatalf("MapAccessInt(1) = (%d, %v), want (20, true)", value, ok)
	}
	if m.used != 1 {
		t.Fatalf("map used = %d, want 1", m.used)
	}
}

func TestMapDeleteIntSmallMap(t *testing.T) {
	m := NewMap(0, true)
	MapInsertInt(m, 1, 10)
	MapInsertInt(m, 2, 20)

	if !MapDeleteInt(m, 1) {
		t.Fatal("MapDeleteInt did not delete an existing key")
	}
	if _, ok := MapAccessInt(m, 1); ok {
		t.Fatal("deleted key is still accessible")
	}
	if m.used != 1 {
		t.Fatalf("map used = %d, want 1", m.used)
	}

	MapInsertInt(m, 3, 30)
	if value, ok := MapAccessInt(m, 3); !ok || value != 30 {
		t.Fatalf("MapAccessInt(3) = (%d, %v), want (30, true)", value, ok)
	}
}

func TestMapPruneTombstones(t *testing.T) {
	m := NewMap(9, true)
	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	table := directory[0]

	keys := make([]int, 0, MapGroupSlots)
	for key := 0; len(keys) < MapGroupSlots; key++ {
		hash := mapHashInt(key, m.seed)
		if uint64(mapH1(hash))&table.groups.lengthMask == 0 {
			keys = append(keys, key)
		}
	}
	for _, key := range keys {
		MapInsertInt(m, key, key+100)
	}

	if !MapDeleteInt(m, keys[0]) || !MapDeleteInt(m, keys[1]) {
		t.Fatal("MapDeleteInt did not create tombstones")
	}
	if !m.tombstonePossible {
		t.Fatal("map did not record possible tombstone")
	}
	if mapCountTombstones(table) != 2 {
		t.Fatalf("tombstones = %d, want 2", mapCountTombstones(table))
	}

	if !mapPruneTombstonesInt(table, m) {
		t.Fatal("mapPruneTombstonesInt did not clean tombstone")
	}
	if mapCountTombstones(table) != 0 {
		t.Fatalf("tombstones after prune = %d, want 0", mapCountTombstones(table))
	}
	if m.tombstonePossible {
		t.Fatal("map still reports possible tombstone after prune")
	}
	for _, key := range keys[2:] {
		value, ok := MapAccessInt(m, key)
		if !ok || value != key+100 {
			t.Fatalf("key %d became inaccessible after prune", key)
		}
	}
}

func TestMapInsertIntGrowsSmallMap(t *testing.T) {
	m := NewMap(0, true)
	for key := range 16 {
		MapInsertInt(m, key, key*10)
	}

	if m.dirLen != 1 {
		t.Fatalf("grown small map dirLen = %d, want 1", m.dirLen)
	}
	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	if directory[0].capacity != 32 {
		t.Fatalf("grown small map capacity = %d, want 32", directory[0].capacity)
	}
	for key := range 16 {
		value, ok := MapAccessInt(m, key)
		if !ok || value != key*10 {
			t.Fatalf("MapAccessInt(%d) = (%d, %v), want (%d, true)", key, value, ok, key*10)
		}
	}
}

func TestMapInsertIntGrowsTable(t *testing.T) {
	m := NewMap(9, true)
	for key := range 15 {
		MapInsertInt(m, key, key+100)
	}

	directory := unsafe.Slice((**MapTable)(m.dirPtr), m.dirLen)
	if directory[0].capacity != 32 {
		t.Fatalf("grown table capacity = %d, want 32", directory[0].capacity)
	}
	for key := range 15 {
		value, ok := MapAccessInt(m, key)
		if !ok || value != key+100 {
			t.Fatalf("MapAccessInt(%d) = (%d, %v), want (%d, true)", key, value, ok, key+100)
		}
	}
}

func TestMapAccessIntTable(t *testing.T) {
	m := NewMap(100, true)
	MapInsertInt(m, 42, 420)

	value, ok := MapAccessInt(m, 42)
	if !ok || value != 420 {
		t.Fatalf("MapAccessInt(42) = (%d, %v), want (420, true)", value, ok)
	}

	if _, ok := MapAccessInt(m, 43); ok {
		t.Fatal("MapAccessInt(43) found an absent key")
	}
}
