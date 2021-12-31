package apiserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xiaodulala/component-tools/pkg/errors"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"socket-demo/internal/apiserver/config"
	"socket-demo/internal/apiserver/controller/home"
	"strconv"
	"strings"
	"time"
)

type apiServer struct {
	*gin.Engine

	healthz    bool
	mode       string
	httpServer *http.Server
	address    string
}

func createApiServer(cfg *config.Config) (*apiServer, error) {
	return &apiServer{
		healthz: cfg.ServerRunOptions.Healthz,
		mode:    cfg.ServerRunOptions.Mode,
		address: net.JoinHostPort(cfg.ApiServerOptions.BindAddress, strconv.Itoa(cfg.ApiServerOptions.BindPort)),
		Engine:  gin.Default(),
	}, nil
}

func (s *apiServer) PrepareRun() *apiServer {
	initRouter(s.Engine)
	home.InitClientHub()
	return s
}

func (s *apiServer) Run() error {
	s.httpServer = &http.Server{
		Addr:    s.address,
		Handler: s,
	}

	var eg errgroup.Group

	eg.Go(func() error {
		logrus.Infof("Start to listening the incoming requests on http address: %s", s.address)

		if err := s.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logrus.Fatal(err.Error())
			return err
		}

		logrus.Infof("Server on %s stopped", s.address)

		return nil
	})

	// Ping the server to make sure the router is working.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if s.healthz {
		if err := s.ping(ctx); err != nil {
			return err
		}
	}

	if err := eg.Wait(); err != nil {
		logrus.Fatal(err.Error())
	}

	return nil
}

func (s *apiServer) ping(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/healthz", s.address)
	if strings.Contains(s.address, "0.0.0.0") {
		url = fmt.Sprintf("http://127.0.0.1:%s/healthz", strings.Split(s.address, ":")[1])
	}

	for {
		// Change NewRequest to NewRequestWithContext and pass context it
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		// Ping the server by sending a GET request to `/healthz`.

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			logrus.Info("The router has been deployed successfully.")

			resp.Body.Close()

			return nil
		}

		// Sleep for a second to continue the next ping.
		logrus.Info("Waiting for the router, retry in 1 second.")
		time.Sleep(1 * time.Second)

		select {
		case <-ctx.Done():
			logrus.Fatal("can not ping http server within the specified time interval.")
		default:
		}
	}
	// return fmt.Errorf("the router has no response, or it might took too long to start up")
}
