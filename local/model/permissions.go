package model

const (
	ServerView   = "server.view"
	ServerCreate = "server.create"
	ServerUpdate = "server.update"
	ServerDelete = "server.delete"
	ServerStart  = "server.start"
	ServerStop   = "server.stop"

	ConfigView   = "config.view"
	ConfigUpdate = "config.update"

	UserView   = "user.view"
	UserCreate = "user.create"
	UserUpdate = "user.update"
	UserDelete = "user.delete"

	RoleView   = "role.view"
	RoleCreate = "role.create"
	RoleUpdate = "role.update"
	RoleDelete = "role.delete"

	MembershipCreate = "membership.create"
	MembershipView   = "membership.view"
	MembershipEdit   = "membership.edit"
)

func AllPermissions() []string {
	return []string{
		ServerView,
		ServerCreate,
		ServerUpdate,
		ServerDelete,
		ServerStart,
		ServerStop,
		ConfigView,
		ConfigUpdate,
		UserView,
		UserCreate,
		UserUpdate,
		UserDelete,
		RoleView,
		RoleCreate,
		RoleUpdate,
		RoleDelete,
		MembershipCreate,
		MembershipView,
		MembershipEdit,
	}
}
