package server

import (
	"fmt"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	STATUS_OK       = "ok"
	STATUS_CREATING = "creating"
	STATUS_DELETING = "deleting"
	STATUS_DELETED  = "deleted"
)

type ServerType struct {
	ID                  primitive.ObjectID  `bson:"_id"`
	Type                string              `json:"type" bson:"type"`
	Name                string              `json:"name" bson:"name"`
	Ip                  string              `json:"ip" bson:"ip"`
	Status              string              `json:"status" bson:"status"`
	UserId              string              `json:userId bson:"user_id"`
	AuthenticationToken string              `json:"authenticationToken" bson:"authentication_token"`
	InstanceId          pulumi.StringOutput `bson:"instance_id"`
	CreatedAt           time.Time           `json:"createdAt" bson:"created_at"`
	PlannedShtudown     time.Time           `json:"plannedShutdown" bson:"planned_shutdown"`
}

type Payload struct {
	Id           string `json:"id"`
	UserId       string `json:"userId"`
	ProductSlug  string `json:"productSlug"`
	ProductId    string `json:"productId"`
	OwnedSeconds int    `json:"ownedSeconds"`
	Amount       string `json:"amount"`
	Reference    string `json:"reference"`
	Timestamp    string `json:"timestamp"`
}

func TypeForProductSlug(slug string) (string, error) {
	if slug == "sniper-small" {
		return "e2-micro", nil
	}

	if slug == "sniper-medium" {
		return "n2-highcpu-8", nil
	}

	if slug == "sniper-big" {
		return "c2-standard-30", nil
	}

	return "", fmt.Errorf("no server found for slug")
}
