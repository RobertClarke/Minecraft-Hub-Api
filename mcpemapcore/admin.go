package mcpemapcore

import "fmt"

func MakeUserAdmin(username string) error {
	var err error
	user, err := LoadUserByUsername(username)

	if user != nil {
		//TODO: get the role from an application global place

		role := Role{}
		role.Id = 1
		role.Name = "Administrator"

		user.AddToRole(role)
		fmt.Printf("User %v is admin\n", user.Username)
	} else {
		fmt.Printf("User %v not found", username)
	}
	return err
}
