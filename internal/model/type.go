package model

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

var (
	googleServers = []*ServerType{
		{
			Slug:         "sniper-small",
			TechnicalKey: "e2-micro",
		},
		{
			Slug:         "sniper-medium",
			TechnicalKey: "n2-highcpu-8",
		},
		{
			Slug:         "sniper-big",
			TechnicalKey: "c2-standard-30",
		},
	}

	// TODO add hetzner servers
	hetznerServers = []*ServerType{
		{
			Slug:         "hetzner-small",
			TechnicalKey: "cx11",
		},
	}
)

type ServerType struct {
	Slug         string `json:"slug" bson:"slug"`
	TechnicalKey string `json:"technicalKey" bson:"technical_key"`
}

type SlugDoesNotBelongToAServerType struct {
	SearchedSlug string
}

func (e *SlugDoesNotBelongToAServerType) Error() string {
	return fmt.Sprintf("slug %s does not belong to a server type", e.SearchedSlug)
}

func ServerTypeForProductSlug(slug string) (*ServerType, error) {

	for _, s := range googleServers {
		if s.Slug == slug {
			return s, nil
		}
	}

	for _, s := range hetznerServers {
		if s.Slug == slug {
			return s, nil
		}
	}

	return nil, &SlugDoesNotBelongToAServerType{SearchedSlug: slug}
}

func (s *ServerType) IsHetznerServer() bool {
	for _, c := range hetznerServers {
		if s.Slug == c.Slug {
			return true
		}
	}

	return false
}

func (s *ServerType) IsGoogleServer() bool {
	for _, c := range googleServers {
		if s.Slug == c.Slug {
			return true
		}
	}

	return false
}

func (s *ServerType) CreationTime() (time.Duration, error) {
	if s.IsGoogleServer() {
		return time.Minute * 1, nil
	}

	if s.IsHetznerServer() {
		return time.Minute * 5, nil
	}

	log.Error().Msgf("can not give a creation time, server type is undefined")
	return 0, fmt.Errorf("can not give a creation time, server type is undefined")
}
