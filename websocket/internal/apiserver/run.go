package apiserver

import "socket-demo/internal/apiserver/config"

func Run(cfg *config.Config) error {
	apiserver, err := createApiServer(cfg)
	if err != nil {
		return err
	}

	return apiserver.PrepareRun().Run()
}
