package storage

import (
	"context"
	"fmt"
	"strings"
)

// ErrStrongholdUnavailable indicates that the formal Stronghold backend cannot
// be opened and callers should decide whether fallback is acceptable.
var ErrStrongholdUnavailable = fmt.Errorf("%w: stronghold backend unavailable", ErrSecretStoreAccessFailed)

// StrongholdSQLiteFallbackProvider uses the current SQLite-backed secret store
// as a fallback implementation while preserving a formal Stronghold lifecycle
// boundary for future production Stronghold wiring.
type StrongholdSQLiteFallbackProvider struct {
	databasePath string
	available    bool
}

// NewStrongholdSQLiteFallbackProvider creates a provider that advertises the
// Stronghold lifecycle contract but falls back to SQLite in development.
func NewStrongholdSQLiteFallbackProvider(databasePath string) *StrongholdSQLiteFallbackProvider {
	return &StrongholdSQLiteFallbackProvider{
		databasePath: strings.TrimSpace(databasePath),
		available:    strings.TrimSpace(databasePath) != "",
	}
}

func (p *StrongholdSQLiteFallbackProvider) Open(ctx context.Context) (SecretStore, error) {
	if p == nil || !p.available || strings.TrimSpace(p.databasePath) == "" {
		return nil, ErrStrongholdUnavailable
	}
	store, err := NewSQLiteSecretStore(p.databasePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStrongholdUnavailable, err)
	}
	select {
	case <-ctx.Done():
		_ = store.Close()
		return nil, ctx.Err()
	default:
	}
	return store, nil
}

func (p *StrongholdSQLiteFallbackProvider) Descriptor() StrongholdDescriptor {
	return StrongholdDescriptor{
		Backend:     "stronghold_sqlite_fallback",
		Available:   p != nil && p.available,
		Fallback:    true,
		Initialized: p != nil && p.available,
	}
}
