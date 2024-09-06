package service

import (
	"log"
	"strconv"

	"github.com/tanush-128/metahost/models"
)

type AppService interface {
	RunApp(conf models.Config) error
}

type appSerivce struct {
	dockerClient DockerClient
}

func NewAppSerivce(dockerClient DockerClient) AppService {
	return &appSerivce{
		dockerClient: dockerClient,
	}
}

func (a *appSerivce) RunApp(conf models.Config) error {
	container, err := a.dockerClient.GetContainer(conf.Name)
	if err != nil {
		return err
	}

	if container.ID != "" {
		if container.Image == conf.Image {
			log.Printf("'%+v' - nothing to modify \n", conf.Name)
			return nil
		} else {
			log.Printf("'%+v' - is being updated \n", conf.Name)

			a.dockerClient.StopContainer(container.ID, &StopContainerOptions{Remove: true})
			a.dockerClient.DeleteImage(container.ImageID)
		}

	} else {
		log.Printf("'%+v' - is being created \n", conf.Name)

	}

	err = a.dockerClient.RunContainer(conf.Image, conf.Name)
	if err != nil {
		log.Println(err)

		return err
	}

	port := strconv.Itoa(conf.Port)

	err = manageServer(conf.Domain, port, "http")
	if err != nil {
		log.Println(err)
		return err
	}

	err = setupSSL(conf.Domain, "tanuedu128@gmail.com")
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
