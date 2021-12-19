package docker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/rjmarques/docker-healthcheck/doctor/model"
)

type Client struct {
	cli  *client.Client
	cont *types.Container
}

func NewClient(dockerHost string, targetContainer string) (*Client, error) {
	cli, err := newDockerClient(dockerHost)
	if err != nil {
		return nil, err
	}

	cont, err := findContainer(cli, targetContainer)
	if err != nil {
		return nil, err
	}

	return &Client{
		cli:  cli,
		cont: cont,
	}, nil
}

func (c *Client) GetHealthChecks() ([]*model.Healthcheck, error) {
	ctx := context.Background()

	res, err := c.cli.ContainerInspect(ctx, c.cont.ID)
	if err != nil {
		return nil, err
	}

	if !model.Status(res.State.Health.Status).Good() {
		return nil, fmt.Errorf("not yet active")
	}

	hc := []*model.Healthcheck{}
	for _, h := range res.State.Health.Log {
		hc = append(hc, toHealthCheck(h))
	}
	return hc, nil
}

func findContainer(cli *client.Client, name string) (*types.Container, error) {
	ctx := context.Background()

	conts, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", name),
		),
	})
	if err != nil {
		return nil, err
	}

	if len(conts) != 1 {
		return nil, fmt.Errorf("failed to find container %s, instead found %d containers", name, len(conts))
	}

	return &conts[0], nil
}

func newDockerClient(dockerHost string) (*client.Client, error) {
	opts, err := genOpt(dockerHost)
	if err != nil {
		return nil, err
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	err = check(cli)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func genOpt(host string) ([]client.Opt, error) {
	helper, err := connhelper.GetConnectionHelper(host)
	if err != nil {
		return nil, err
	}

	if helper == nil {
		return []client.Opt{
			client.WithAPIVersionNegotiation(),
		}, nil
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}
	return []client.Opt{
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithDialContext(helper.Dialer),
		client.WithAPIVersionNegotiation(),
	}, nil
}

func check(cli *client.Client) error {
	// check if the connection is working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := cli.Ping(ctx)
	return err
}

func toHealthCheck(h *types.HealthcheckResult) *model.Healthcheck {
	return &model.Healthcheck{
		ID:       h.Output,
		Start:    h.Start,
		End:      h.End,
		ExitCode: h.ExitCode,
	}
}
