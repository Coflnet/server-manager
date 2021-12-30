package mongo

import (
	"context"
	"math/rand"
	"server-manager/server"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Create(server *server.ServerType) (*server.ServerType, error) {
	log.Info().Msgf("inserting server %s into database", server.Type)

	server.ID = primitive.NewObjectID()
	server.Name = createRandomString(10)
	server.AuthenticationToken = createRandomString(20)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	res, err := serverCollection.InsertOne(ctx, server)

	if err != nil {
		log.Error().Err(err).Msgf("error when inserting to mongo db")
		return nil, err
	}

	result := serverCollection.FindOne(ctx, bson.M{"_id": res.InsertedID})

	result.Decode(server)

	return server, nil

}

func List() ([]*server.ServerType, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, err := serverCollection.Find(ctx, bson.D{})
	var servers []*server.ServerType
	if err != nil {
		log.Error().Err(err).Msgf("erorr when selecting servers")
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result server.ServerType
		err := cur.Decode(&result)
		if err != nil {
			log.Error().Err(err).Msgf("erorr when decoding server")
			return nil, err
		}
		servers = append(servers, &result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func Delete(server *server.ServerType) error {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := serverCollection.DeleteOne(ctx, bson.M{"_id": server.ID})

	return err
}

func Update(server *server.ServerType) (*server.ServerType, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err := serverCollection.UpdateOne(
		ctx,
		bson.M{"_id": server.ID},
		bson.D{
			{"$set", bson.D{{"status", server.Status}}},
			{"$set", bson.D{{"planned_shutdown", server.PlannedShtudown}}},
			{"$set", bson.D{{"ip", server.Ip}}},
		},
	)

	if err != nil {
		return nil, err
	}

	result := serverCollection.FindOne(ctx, bson.M{"_id": server.ID})

	result.Decode(server)

	return server, nil
}

func createRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
