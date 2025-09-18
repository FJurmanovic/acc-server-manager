package repository

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/graceful"
	"acc-server-manager/local/utl/logging"
	"context"
	"time"

	"go.uber.org/dig"
)

// *dig.Container: Dig Container
func InitializeRepositories(c *dig.Container) {
	c.Provide(NewServiceControlRepository)
	c.Provide(NewStateHistoryRepository)
	c.Provide(NewServerRepository)
	c.Provide(NewConfigRepository)
	c.Provide(NewLookupRepository)
	c.Provide(NewSteamCredentialsRepository)
	c.Provide(NewMembershipRepository)

	if err := c.Provide(func() *model.Steam2FAManager {
		manager := model.NewSteam2FAManager()

		shutdownManager := graceful.GetManager()
		shutdownManager.RunGoroutine(func(ctx context.Context) {
			ticker := time.NewTicker(15 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					manager.CleanupOldRequests(30 * time.Minute)
				}
			}
		})

		return manager
	}); err != nil {
		logging.Panic("unable to initialize steam 2fa manager")
	}
}
