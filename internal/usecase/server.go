package usecase

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"math/rand"
	"server-manager/internal/kafka"
	"server-manager/internal/model"
	"server-manager/internal/mongo"
	"time"
)

// calculatePlannedShutdown calculates the planned shutdown date for a server
// based on the server type a different time is used for each server type
// hetzner servers have a longer time to create than google servers,
// therefore the user should get a longer time to use it
func calculatePlannedShutdown(t *model.ServerType, d time.Duration) (time.Time, error) {
	result := time.Now()

	result = result.Add(d)
	log.Debug().Msgf("adding %d minutes to planned shutdown as owned time; res = %s", d.Minutes(), result.Format(time.RFC3339))

	creationTime, err := t.CreationTime()
	if err != nil {
		return time.Now(), err
	}

	result = result.Add(creationTime)
	log.Debug().Msgf("add a creation time of %d minutes to the planned shutdown; res = %s", creationTime.Minutes(), result.Format(time.RFC3339))

	return result, nil
}

func createAuthenticationTokenForServer() string {
	return RandString(21)
}

func createServerName(t *model.ServerType, userId string) string {
	n := time.Now()
	formatted := n.Format("15-04-05")

	return fmt.Sprintf("server-%s-%s-%s", t.Slug, userId, formatted)
}

func serverStateChanged(s *model.Server) error {

	var errCh = make(chan error)

	// persist state in mongo
	go func(ch chan error) {
		ch <- mongo.UpdateStatus(s)
	}(errCh)

	// publish a kafka message
	go func(ch chan error) {
		ch <- kafka.ProduceServerState(s)
	}(errCh)

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			log.Error().Err(err).Msgf("error when setting server %s to state %s", s.Name, s.Status)
			return err
		}
	}

	log.Debug().Msgf("set the state %s for server %s", s.Status, s.Name)

	return nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func RandString(n int) string {

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const (
		letterIdxBits = 6
		letterIdxMask = 1<<letterIdxBits - 1
	)

	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}
