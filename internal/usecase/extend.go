package usecase

import (
	"fmt"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"time"
)

func ExtendServer(server *model.Server, duration time.Duration) error {

	if duration.Seconds() < 0 {
		return fmt.Errorf("duration must be positive")
	}

	currentShutdown := server.PlannedShutdown

	newShutdown := currentShutdown.Add(duration)
	timeInOneDay := time.Now().Add(time.Hour * 24)

	if newShutdown.After(timeInOneDay) {
		return fmt.Errorf("new shutdown time is more than 24 hours in the future, %s", newShutdown.Format(time.RFC3339))
	}

	server.PlannedShutdown = &newShutdown
	err := mongo.UpdatePlannedShutdown(server)
	if err != nil {
		return err
	}

	return nil
}
