package userdetails

import (
	"encoding/json"
	"fmt"

	context "golang.org/x/net/context"
)

// UserDetails implements the GreeterServer for tests
type UserDetails struct{}

// GetUserDetails implements userdetails.UserDetailsServer
func (s *UserDetails) GetUserDetails(ctx context.Context, in *UserDetailsRequest) (*UserDetailsReply, error) {
	// User details stored on server/database
	detailsOnServer := map[string]interface{}{
		"name":    "Anuj",
		"age":     26,
		"married": false,
		"city":    []string{"Mumbai", "Bangalore"},
		"games":   []interface{}{"Crisis", 2048, "Contra"},
	}

	var userQuery map[string]interface{}
	err := json.Unmarshal(in.Properties, &userQuery)
	if err != nil {
		fmt.Println("Could not unmarshal user query from request")
		return nil, err
	}

	var response = make(map[string]interface{})
	for key, val := range userQuery {
		if detailsOnServer[key] != nil {
			// Add the property (present on server) in the reply
			response[key] = detailsOnServer[key]
		} else {
			// Give default value (sent by client) if property not on server
			response[key] = val
		}
	}

	serializedResponse, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Could not send back user details in reply")
		return nil, err
	}

	return &UserDetailsReply{Details: serializedResponse}, nil
}

// NewUserDetails creates new UserDetails server
func NewUserDetails() *UserDetails {
	return &UserDetails{}
}
