package mcpemapcore

type Backend interface {
	CreateMap(user *User,
		newMap *NewMap)
}
