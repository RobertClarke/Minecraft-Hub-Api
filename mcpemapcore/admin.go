package mcpemapcore

import "fmt"

func MakeUserAdmin(username string) error {
	var err error
	user, err := LoadUserByUsername(username)

	if user != nil {
		user.AddToRole(*rolesByName["Administrator"])
		fmt.Printf("User %v is admin\n", user.Username)
	} else {
		fmt.Printf("User %v not found", username)
	}
	return err
}
