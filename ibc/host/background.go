package host

import (
	"context"
	"time"

	"github.com/pokt-network/pocket/shared/modules"
)

// StartBackgroundTasks starts the background tasks managed by the IBC host
func (h *ibcHost) StartBackgroundTasks(ctx context.Context) error {
	go h.flushCache(ctx, h.GetBus())
	return nil
}

// flushCache flushes the cache according to the flush interval set in the host configuration
func (h *ibcHost) flushCache(ctx context.Context, bus modules.Bus) {
	bsc := bus.GetBulkStoreCacher()
	ticker := time.NewTicker(time.Duration(h.cfg.BulkStoreCacher.FlushIntervalSeconds) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := bsc.FlushAllEntries(); err != nil {
					h.logger.Error().Err(err).Msg("🚨 Error Flushing Bulk Store Cacher 🚨")
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
