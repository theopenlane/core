package config

import (
	"sync"
	"testing"
	"time"
)

type dummyProvider struct {
	m     sync.Mutex
	cfg   *Config
	calls int
}

func (d *dummyProvider) Get() (*Config, error) {
	d.m.Lock()
	defer d.m.Unlock()
	d.calls++
	return d.cfg, nil
}

func TestProviderWithRefresh_NoRefresh(t *testing.T) {
	t.Parallel()

	dp := &dummyProvider{cfg: &Config{}}
	pr, err := NewProviderWithRefresh(dp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := pr.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dp.cfg {
		t.Fatal("expected same config")
	}
	if dp.calls != 1 {
		t.Fatalf("expected 1 call, got %d", dp.calls)
	}
	pr.Close()
}

func TestProviderWithRefresh_Refresh(t *testing.T) {
	t.Parallel()

	base := &Config{}
	base.Settings.RefreshInterval = 10 * time.Millisecond
	dp := &dummyProvider{cfg: base}
	pr, err := NewProviderWithRefresh(dp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// change config after creation
	dp.m.Lock()
	dp.cfg = &Config{}
	dp.m.Unlock()
	time.Sleep(1000 * time.Millisecond)
	dp.m.Lock()
	c := dp.calls
	dp.m.Unlock()
	if c < 2 {
		t.Fatalf("expected provider to be called at least twice, got %d", c)
	}
	pr.Close()
}
