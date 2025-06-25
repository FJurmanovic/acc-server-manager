package repository

import (
	"go.uber.org/dig"
)

// InitializeRepositories
// Initializes Dependency Injection modules for repositories
//
//	Args:
//		*dig.Container: Dig Container
func InitializeRepositories(c *dig.Container) {
	c.Provide(NewApiRepository)
	c.Provide(NewStateHistoryRepository)
	c.Provide(NewServerRepository)
	c.Provide(NewConfigRepository)
	c.Provide(NewLookupRepository)
	c.Provide(NewSteamCredentialsRepository)
	c.Provide(NewSystemConfigRepository)
	c.Provide(NewMembershipRepository)
}
