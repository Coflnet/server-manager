package mongo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var serverCollection *mongo.Collection

func Connect() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_HOST")))

	if err != nil {
		return err
	}

	serverCollection = client.Database("server_manager").Collection("servers")

	log.Info().Msgf("using mongo host %s", os.Getenv("MONGO_HOST"))

	return nil
}

func Disconnect() error {
	if client == nil {
		fmt.Errorf("client is nil, therefore can not disconnect")
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client.Disconnect(ctx)

	return nil
}
