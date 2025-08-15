package cache_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"weather-service/internal/cache"
)

func TestSetGet(t *testing.T) {
	c := cache.NewCache(100 * time.Millisecond)
	c.Set("k", "v")
	v, ok := c.Get("k")
	if !ok {
		t.Fatalf("expected key to be present")
	}
	if s, ok2 := v.(string); !ok2 || s != "v" {
		t.Fatalf("got %v (%T), want 'v' (string)", v, v)
	}
}

func TestExpiration(t *testing.T) {
	c := cache.NewCache(10 * time.Millisecond)
	c.Set("foo", 42)
	// Should be present immediately
	if _, ok := c.Get("foo"); !ok {
		t.Fatalf("expected key to be present before TTL")
	}
	time.Sleep(15 * time.Millisecond)
	if v, ok := c.Get("foo"); ok {
		t.Fatalf("expected key to be expired, got %v", v)
	}
	// After expiration Get also deletes the entry; setting again should work
	c.Set("foo", 100)
	if v, ok := c.Get("foo"); !ok || v.(int) != 100 {
		t.Fatalf("expected re-set key to be present with value 100, got (%v, %v)", v, ok)
	}
}

func TestDelete(t *testing.T) {
	c := cache.NewCache(50 * time.Millisecond)
	c.Set("del", true)
	if _, ok := c.Get("del"); !ok {
		t.Fatalf("expected key to exist before delete")
	}
	c.Delete("del")
	if v, ok := c.Get("del"); ok {
		t.Fatalf("expected key to be deleted, got %v", v)
	}
}

func TestOverwriteResetsTTLAndValue(t *testing.T) {
	c := cache.NewCache(30 * time.Millisecond)
	c.Set("k", "v1")
	time.Sleep(15 * time.Millisecond)
	c.Set("k", "v2")
	// Immediately after overwrite, value should be updated
	if v, ok := c.Get("k"); !ok || v.(string) != "v2" {
		t.Fatalf("expected updated value 'v2', got (%v, %v)", v, ok)
	}
	// Wait less than the second TTL to ensure it's still present
	time.Sleep(20 * time.Millisecond)
	if v, ok := c.Get("k"); !ok || v.(string) != "v2" {
		t.Fatalf("expected value to remain before new TTL expiry, got (%v, %v)", v, ok)
	}
	// Now wait enough for the second TTL to pass and ensure expiration
	time.Sleep(15 * time.Millisecond)
	if v, ok := c.Get("k"); ok {
		t.Fatalf("expected key to be expired after second TTL, got %v", v)
	}
}

func TestZeroTTLExpiresImmediately(t *testing.T) {
	c := cache.NewCache(0)
	c.Set("k", 1)
	// Even an immediate Get may race the exact same timestamp; allow a tiny sleep
	time.Sleep(1 * time.Millisecond)
	if v, ok := c.Get("k"); ok {
		t.Fatalf("expected zero-TTL entry to be expired, got %v", v)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := cache.NewCache(100 * time.Millisecond)
	const goroutines = 10
	const perG = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				key := "k" + strconv.Itoa(g) + ":" + strconv.Itoa(i)
				c.Set(key, i)
				if v, ok := c.Get(key); !ok || v.(int) != i {
					t.Errorf("get after set failed for %s: got (%v,%v)", key, v, ok)
				}
			}
		}()
	}
	wg.Wait()
	// Spot check a few keys still present before TTL
	for g := 0; g < goroutines; g++ {
		key := "k" + strconv.Itoa(g) + ":" + strconv.Itoa(perG-1)
		if _, ok := c.Get(key); !ok {
			t.Fatalf("expected key %s to be present", key)
		}
	}
}
