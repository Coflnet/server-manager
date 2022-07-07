package mongo

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client           *mongo.Client
	serverCollection *mongo.Collection
)

func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_HOST")))

	if err != nil {
		log.Panic().Err(err).Msg("can not connect to mongo")
	}

	db := client.Database("server_manager")
	serverCollection = db.Collection("servers")
}

func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Warn().Msg("could not disconnect from mongo")
	}
}
