package usecase

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// CreateStateTransferToken gets a new state token from the sniper service
func CreateStateTransferToken() (string, error) {
	request, err := http.NewRequest(http.MethodGet, sniperUrl(), nil)
	if err != nil {
		return "", err
	}

	client := http.DefaultClient
	client.Timeout = time.Second * 10

	log.Info().Msgf("using path: %s", sniperUrl())

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error().Err(err).Msgf("error requesting a new sniper transfer token")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic().Err(err).Msgf("somehow body could not been read")
	}

	str := string(b)

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("error requesting a new sniper transfer token, status code %d\n%s", resp.StatusCode, str)
	}

	log.Info().Msgf("received a new sniper transfer token: %s, status code: %v, body: %v", str, resp.Status, b)

	return str, nil
}

func sniperUrl() string {
	baseUrl := os.Getenv("SNIPER_BASE_URL")

	if baseUrl == "" {
		log.Panic().Msg("SNIPER_BASE_URL is not set")
	}

	return fmt.Sprintf("%sapi/Sniper/token", baseUrl)
}
