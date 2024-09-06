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
	nginxSerivce NginxService
	sslService   SSLService
}

func NewAppSerivce(dockerClient DockerClient,
	nginxSerivce NginxService,
	sslService SSLService,
) AppService {
	return &appSerivce{
		dockerClient: dockerClient,
		nginxSerivce: nginxSerivce,
		sslService:   sslService,
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

	port := strconv.Itoa(conf.Port)

	err = a.dockerClient.RunContainer(conf.Image, conf.Name, port, port)
	if err != nil {
		log.Println(err)

		return err
	}

	exists, err := a.nginxSerivce.CheckIfServerExists(conf.Domain, port, "http")
	if err != nil {
		log.Println(err)
		return err
	}

	if exists {
		log.Printf("'%+v' - Nginx configuration already exists \n", conf.Domain)
		err = a.sslService.SetupSSL(conf.Domain, "tanuedu128@gmail.com")
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}

	err = a.nginxSerivce.AddServer(conf.Domain, port, "http")
	if err != nil {
		log.Println(err)
		return err
	}

	err = a.sslService.SetupSSL(conf.Domain, "tanuedu128@gmail.com")
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
