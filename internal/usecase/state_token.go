package usecase

import (
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error().Err(err).Msgf("error requesting a new sniper transfer token")
	}

	str, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic().Err(err).Msgf("somehow body could not been read")
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("error requesting a new sniper transfer token, status code %d\n%s", resp.StatusCode, str)
	}

	return string(str), nil
}

func sniperUrl() string {
	baseUrl := os.Getenv("SNIPER_BASE_URL")

	if baseUrl == "" {
		log.Panic().Msg("SNIPER_BASE_URL is not set")
	}

	return path.Join(baseUrl, "/api/Sniper/token")
}
