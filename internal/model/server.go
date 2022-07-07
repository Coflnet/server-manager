package model

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ServerStatusOk       = "ok"
	ServerStatusCreating = "creating"
	ServerStatusCreated  = "created"
	ServerStatusDeleting = "deleting"
	ServerStatusDeleted  = "deleted"
	ServerStatusPlanned  = "planned"
)

type Server struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `json:"name" bson:"name"`
	Type *ServerType        `json:"type" bson:"type"`

	Ip                  string               `json:"ip" bson:"ip"`
	Status              string               `json:"status" bson:"status"`
	UserId              string               `json:"userId" bson:"user_id"`
	AuthenticationToken string               `json:"authenticationToken" bson:"authentication_token"`
	InstanceId          *pulumi.StringOutput `bson:"instance_id"`
	CreatedAt           *time.Time           `json:"createdAt" bson:"created_at"`
	PlannedShutdown     *time.Time           `json:"plannedShutdown" bson:"planned_shutdown"`
}

type ServerInvalidError struct {
	Reason string
	Server *Server
}

func (e *ServerInvalidError) Error() string {
	return fmt.Sprintf("the server %s is invalid, reason is: %s", e.Server.Name, e.Reason)
}

func FilterGoogleServers(servers []*Server) []*Server {
	var filteredServers []*Server

	for _, s := range servers {
		if s.Type.IsGoogleServer() {
			filteredServers = append(filteredServers, s)
		}
	}

	return filteredServers
}

func FilterHetznerServers(servers []*Server) []*Server {
	var filteredServers []*Server

	for _, s := range servers {
		if s.Type.IsHetznerServer() {
			filteredServers = append(filteredServers, s)
		}
	}

	return filteredServers
}

func IsServer100Gbit(typeOfServer string) bool {
	if typeOfServer == "n2-highcpu-80" {
		return true
	}

	if typeOfServer == "c2-standard-60" {
		return true
	}

	return false
}
