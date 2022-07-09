package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"server-manager/internal/model"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
)

func ListActiveServers() ([]*model.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.D{{"$or", bson.A{
		bson.M{"status": model.ServerStatusCreating},
		bson.M{"status": model.ServerStatusCreated},
		bson.M{"status": model.ServerStatusOk},
	}}}

	cur, err := serverCollection.Find(ctx, filter)

	var servers []*model.Server

	if err != nil {
		log.Error().Err(err).Msgf("error when querying servers")
		return nil, err
	}

	defer func(cur *mongo.Cursor, ctx context.Context) {
		err := cur.Close(ctx)
		if err != nil {
			log.Warn().Msg("error closing server collection cursor")
		}
	}(cur, ctx)

	err = cur.All(ctx, &servers)
	if err != nil {
		log.Error().Err(err).Msgf("error deserializing servers")
		return nil, err
	}

	return servers, nil
}

func ServerByName(name string) (*model.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var server model.Server
	err := serverCollection.FindOne(ctx, bson.M{"name": name}).Decode(&server)

	if err != nil {
		log.Error().Err(err).Msgf("error when querying server %s", name)
		return nil, err
	}

	return &server, nil
}

func InsertServer(server *model.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := serverCollection.InsertOne(ctx, server)
	if err != nil {
		log.Error().Err(err).Msgf("error when inserting server %s", server.Name)
		return err
	}

	if res.InsertedID == nil {
		log.Warn().Msgf("server was inserted but id is nil")
	}

	return nil
}

func UpdateIp(server *model.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := serverCollection.UpdateOne(ctx, bson.M{"_id": server.ID}, bson.M{"$set": bson.M{"ip": server.Ip}})

	if err != nil {
		log.Error().Err(err).Msgf("error when updating ip from server %v %s", server.ID, server.Name)
		return err
	}

	if res.ModifiedCount == 0 {
		log.Warn().Msgf("server %v %s not found", server.ID, server.Name)
	}

	return nil
}

func UpdateStatus(server *model.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := serverCollection.UpdateOne(ctx, bson.M{"_id": server.ID}, bson.M{"$set": bson.M{"status": server.Status}})

	if err != nil {
		log.Error().Err(err).Msgf("error when updating status for server %v %s", server.ID, server.Name)
		return err
	}

	if res.ModifiedCount == 0 {
		log.Warn().Msgf("server %v %s not found", server.ID, server.Name)
	}

	return nil
}

func UpdatePlannedShutdown(server *model.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := serverCollection.UpdateOne(ctx, bson.M{"_id": server.ID}, bson.M{"$set": bson.M{"planned_shutdown": server.PlannedShutdown}})

	if err != nil {
		log.Error().Err(err).Msgf("error when updating planned shutdown for server %v %s", server.ID, server.Name)
		return err
	}

	if res.ModifiedCount == 0 {
		log.Warn().Msgf("server %v %s not found", server.ID, server.Name)
	}

	return nil
}
