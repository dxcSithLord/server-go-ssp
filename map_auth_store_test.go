package ssp

import "testing"

func TestMapAuthStore_SaveAndFind(t *testing.T) {
	store := NewMapAuthStore()
	identity := &SqrlIdentity{
		Idk:  "test-idk-123",
		Suk:  "test-suk",
		Vuk:  "test-vuk",
		Pidk: "test-pidk",
	}

	err := store.SaveIdentity(identity)
	if err != nil {
		t.Fatalf("SaveIdentity failed: %v", err)
	}

	found, err := store.FindIdentity("test-idk-123")
	if err != nil {
		t.Fatalf("FindIdentity failed: %v", err)
	}
	if found.Idk != "test-idk-123" {
		t.Errorf("Expected idk test-idk-123, got %s", found.Idk)
	}
	if found.Suk != "test-suk" {
		t.Errorf("Expected suk test-suk, got %s", found.Suk)
	}
	if found.Vuk != "test-vuk" {
		t.Errorf("Expected vuk test-vuk, got %s", found.Vuk)
	}
}

func TestMapAuthStore_FindNotFound(t *testing.T) {
	store := NewMapAuthStore()

	_, err := store.FindIdentity("nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMapAuthStore_Delete(t *testing.T) {
	store := NewMapAuthStore()
	identity := &SqrlIdentity{
		Idk: "delete-me",
		Suk: "suk",
		Vuk: "vuk",
	}

	err := store.SaveIdentity(identity)
	if err != nil {
		t.Fatalf("SaveIdentity failed: %v", err)
	}

	// Verify it exists
	_, err = store.FindIdentity("delete-me")
	if err != nil {
		t.Fatalf("FindIdentity failed before delete: %v", err)
	}

	err = store.DeleteIdentity("delete-me")
	if err != nil {
		t.Fatalf("DeleteIdentity failed: %v", err)
	}

	// Verify it's gone
	_, err = store.FindIdentity("delete-me")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestMapAuthStore_DeleteNonexistent(t *testing.T) {
	store := NewMapAuthStore()
	// Should not error when deleting something that doesn't exist
	err := store.DeleteIdentity("nonexistent")
	if err != nil {
		t.Errorf("Expected no error deleting nonexistent identity, got %v", err)
	}
}

func TestMapAuthStore_SaveOverwrite(t *testing.T) {
	store := NewMapAuthStore()

	identity1 := &SqrlIdentity{
		Idk: "overwrite-test",
		Suk: "original-suk",
	}
	err := store.SaveIdentity(identity1)
	if err != nil {
		t.Fatalf("SaveIdentity failed: %v", err)
	}

	identity2 := &SqrlIdentity{
		Idk: "overwrite-test",
		Suk: "updated-suk",
	}
	err = store.SaveIdentity(identity2)
	if err != nil {
		t.Fatalf("SaveIdentity (overwrite) failed: %v", err)
	}

	found, err := store.FindIdentity("overwrite-test")
	if err != nil {
		t.Fatalf("FindIdentity failed: %v", err)
	}
	if found.Suk != "updated-suk" {
		t.Errorf("Expected updated-suk, got %s", found.Suk)
	}
}

func TestMapAuthStore_MultipleIdentities(t *testing.T) {
	store := NewMapAuthStore()

	for i := 0; i < 10; i++ {
		identity := &SqrlIdentity{
			Idk: Nut("idk-" + string(rune('A'+i))).String(),
			Suk: "suk",
		}
		if err := store.SaveIdentity(identity); err != nil {
			t.Fatalf("SaveIdentity %d failed: %v", i, err)
		}
	}

	// Verify they all exist
	for i := 0; i < 10; i++ {
		idk := Nut("idk-" + string(rune('A'+i))).String()
		_, err := store.FindIdentity(idk)
		if err != nil {
			t.Errorf("FindIdentity for %s failed: %v", idk, err)
		}
	}
}

func TestMapAuthStore_DisabledIdentity(t *testing.T) {
	store := NewMapAuthStore()
	identity := &SqrlIdentity{
		Idk:      "disabled-user",
		Suk:      "suk",
		Vuk:      "vuk",
		Disabled: true,
	}

	err := store.SaveIdentity(identity)
	if err != nil {
		t.Fatalf("SaveIdentity failed: %v", err)
	}

	found, err := store.FindIdentity("disabled-user")
	if err != nil {
		t.Fatalf("FindIdentity failed: %v", err)
	}
	if !found.Disabled {
		t.Error("Expected identity to be disabled")
	}
}

func (n Nut) String() string {
	return string(n)
}

func BenchmarkMapAuthStore_Save(b *testing.B) {
	store := NewMapAuthStore()
	identity := &SqrlIdentity{Idk: "bench-idk", Suk: "suk", Vuk: "vuk"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.SaveIdentity(identity)
	}
}

func BenchmarkMapAuthStore_Find(b *testing.B) {
	store := NewMapAuthStore()
	identity := &SqrlIdentity{Idk: "bench-idk", Suk: "suk", Vuk: "vuk"}
	_ = store.SaveIdentity(identity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.FindIdentity("bench-idk")
	}
}
