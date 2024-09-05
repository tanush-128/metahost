package service

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	containertypes "github.com/docker/docker/api/types/container"
)

type DockerClient interface {
	RunContainer(imageName string, containerName string) error
	StopContainer(id string, opts *StopContainerOptions) error
	DeleteImage(id string) error
	GetContainer(name string) (types.Container, error)
	// DeleteImage()
}

type dockerClient struct {
	ctx context.Context
	cli *client.Client
}

type StopContainerOptions struct {
	//If true then continer is removed after stoping
	Remove bool
}

func NewDockerClient(ctx context.Context, cli *client.Client) DockerClient {
	return &dockerClient{
		ctx: ctx,
		cli: cli,
	}
}

func (dc *dockerClient) GetContainer(containerName string) (types.Container, error) {
	containers, err := dc.cli.ContainerList(dc.ctx, containertypes.ListOptions{All: true})
	if err != nil {

		return types.Container{}, err
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				return container, nil
			}
		}
	}

	return types.Container{}, nil
}

func (dc *dockerClient) RunContainer(imageName string, containerName string) error {

	out, err := dc.cli.ImagePull(dc.ctx, imageName, image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(io.Discard, out)

	port, _ := nat.NewPort("tcp", "8080")
	resp, err := dc.cli.ContainerCreate(dc.ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "8000",
				},
			},
		},
	}, nil, nil, containerName)
	if err != nil {
		return err
	}

	if err := dc.cli.ContainerStart(dc.ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (dc *dockerClient) StopContainer(id string, opts *StopContainerOptions) error {
	noWaitTimeout := 0 // to not wait for the container to exit gracefully

	if err := dc.cli.ContainerStop(dc.ctx, id, containertypes.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		return err
	}
	if opts.Remove {
		err := dc.cli.ContainerRemove(dc.ctx, id, containertypes.RemoveOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}

func (dc *dockerClient) DeleteImage(id string) error {
	_, err := dc.cli.ImageRemove(dc.ctx, id, image.RemoveOptions{})
	if err != nil {
		return err
	}

	// fmt.Println("deleting")

	return nil
}
