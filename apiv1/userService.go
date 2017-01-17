package main

import "net/http"

// User structure
type User struct {
	ID       string
	Username string `redis:"username"`
	Auth     string `redis:"auth"`
}

// UserService interface
type UserService interface {
	GetUser(wr http.ResponseWriter, r *http.Request) *User
}

// implementation of userService
type userService struct {
	currentBackend Backend
}

// CreateGetUserService gets a new instance of the UserService
func CreateGetUserService() UserService {
	service := userService{}
	service.currentBackend = mySQLBackend{}
	return service
}

func (s userService) GetUser(wr http.ResponseWriter, r *http.Request) *User {
	userid := r.Header.Get("userid")
	//TODO switch to production auth backend
	user, err := s.currentBackend.LoadUserInfo(userid)
	if hasFailed(wr, err) {
		return nil
	}
	return user
}
