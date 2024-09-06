package main

import (
	"context"
	"log"
	"os"

	docker_client "github.com/docker/docker/client"
	"github.com/google/go-github/v64/github"
	"github.com/joho/godotenv"
	"github.com/tanush-128/metahost/service"
	"golang.org/x/oauth2"
)

type RepoConfig struct {
	Owner string
	Name  string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Unable to load env")
	}

	github_access_token := os.Getenv("GITHUB_API_KEY")

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: github_access_token},
	)
	tc := oauth2.NewClient(ctx, ts)
	// get go-github client
	client := github.NewClient(tc)

	repoConfig := RepoConfig{
		Owner: "tanush-128",
		Name:  "metaserver-conf",
	}

	cli, err := docker_client.NewClientWithOpts(docker_client.FromEnv, docker_client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	dockerClient := service.NewDockerClient(ctx, cli)
	nginxService := service.NewNginxSerivce(dockerClient)
	sslService := service.NewSSLService()

	err = nginxService.UpdateOrInstallNginx()
	if err != nil {
		panic(err)
	}

	appService := service.NewAppSerivce(dockerClient, nginxService, sslService)

	// content, err := ioutil.ReadFile("test.yaml")
	// testAppConf, err := utils.ParseYaml(string(content))
	// if err != nil {
	// 	panic(err)
	// }

	// err = appService.RunApp(testAppConf)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// dockerClient.RunContainer(testAppConf.Image, testAppConf.Name)
	// RunContainer(ctx, cli, testConf)

	// GetFiles(client, repoConfig)

	monitorRepo(ctx, client, repoConfig, appService)

}
