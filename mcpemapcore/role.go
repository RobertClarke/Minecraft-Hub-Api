package mcpemapcore

type Role struct {
	Id   int
	Name string
}

func GetRole(name string) Role {
	return *rolesByName[name]
}
