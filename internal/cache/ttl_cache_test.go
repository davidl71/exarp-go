package cache

import (
	"testing"
	"time"
)

func TestTTLCache_GetSet(t *testing.T) {
	c := NewTTLCache()
	key := "k1"
	val := []byte("v1")

	c.Set(key, val, 5*time.Minute)
	got, ok := c.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if string(got) != string(val) {
		t.Errorf("got %q, want %q", got, val)
	}
}

func TestTTLCache_GetMiss(t *testing.T) {
	c := NewTTLCache()
	got, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestTTLCache_Expiry(t *testing.T) {
	c := NewTTLCache()
	key := "k1"
	c.Set(key, []byte("v1"), 10*time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	got, ok := c.Get(key)
	if ok {
		t.Fatal("expected miss after TTL")
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestTTLCache_Invalidate(t *testing.T) {
	c := NewTTLCache()
	key := "k1"
	c.Set(key, []byte("v1"), 5*time.Minute)
	c.Invalidate(key)
	got, ok := c.Get(key)
	if ok {
		t.Fatal("expected miss after Invalidate")
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestTTLCache_Clear(t *testing.T) {
	c := NewTTLCache()
	c.Set("k1", []byte("v1"), 5*time.Minute)
	c.Set("k2", []byte("v2"), 5*time.Minute)
	c.Clear()
	if _, ok := c.Get("k1"); ok {
		t.Error("k1 should be cleared")
	}
	if _, ok := c.Get("k2"); ok {
		t.Error("k2 should be cleared")
	}
}

func TestTTLCache_SetZeroTTL(t *testing.T) {
	c := NewTTLCache()
	c.Set("k1", []byte("v1"), 0)
	_, ok := c.Get("k1")
	if ok {
		t.Error("Set with zero TTL should not store")
	}
}

func TestGetScorecardCache(t *testing.T) {
	c1 := GetScorecardCache()
	c2 := GetScorecardCache()
	if c1 != c2 {
		t.Error("GetScorecardCache should return same instance")
	}
	c1.Set("test", []byte("x"), time.Minute)
	got, ok := c2.Get("test")
	if !ok || string(got) != "x" {
		t.Errorf("shared cache: got ok=%v, want true; got %q", ok, got)
	}
}
