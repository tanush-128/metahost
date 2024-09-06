package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	docker_client "github.com/docker/docker/client"
	"github.com/google/go-github/v64/github"
	"github.com/joho/godotenv"
	"github.com/tanush-128/metahost/models"
	"github.com/tanush-128/metahost/service"
	"github.com/tanush-128/metahost/utils"
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

func monitorRepo(ctx context.Context, client *github.Client, repoConfig RepoConfig, appService service.AppService) {
	var latestCommit *github.RepositoryCommit
	for {
		commits, _, err := client.Repositories.ListCommits(ctx, repoConfig.Owner, repoConfig.Name, nil)
		if err != nil {
			log.Println(err)
		}
		if latestCommit == nil {
			latestCommit = commits[0]
			fmt.Printf("last found commit : %s \n", *latestCommit.Commit.Message)
			appConfs, err := GetAppConfs(client, repoConfig)
			if err != nil {
				fmt.Println(err)
			}

			for _, appConf := range appConfs {
				go appService.RunApp(appConf)
			}

		} else if *latestCommit.SHA != *commits[0].SHA {
			latestCommit = commits[0]

			log.Printf("new found commit : %s \n", *latestCommit.Commit.Message)

			appConfs, err := GetAppConfs(client, repoConfig)
			if err != nil {
				fmt.Println(err)
			}

			for _, appConf := range appConfs {
				go appService.RunApp(appConf)
			}
		} else {

			log.Println("No New Commit Found")
		}

		time.Sleep(60 * time.Second)
	}
}

func GetAppConfs(client *github.Client, repoConfig RepoConfig) ([]models.Config, error) {
	ctx := context.Background()
	file, dirs, _, _ := client.Repositories.GetContents(ctx, repoConfig.Owner, repoConfig.Name, "", nil)

	if file != nil {

		fmt.Print(file.Content)
	}

	var confs []models.Config

	for _, dir := range dirs {
		file, _, _, _ := client.Repositories.GetContents(ctx, repoConfig.Owner, repoConfig.Name, dir.GetPath(), nil)
		fileType := strings.Split(file.GetName(), ".")[1]
		if fileType == "yaml" || fileType == "yml" {
			data, err := file.GetContent()
			if err != nil {
				return confs, err
			}
			conf, err := utils.ParseYaml(data)
			if err != nil {
				return confs, err
			}
			fmt.Printf("%+v \n", conf)
			confs = append(confs, conf)
		}
	}

	return confs, nil
}
