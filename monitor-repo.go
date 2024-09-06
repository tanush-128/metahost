package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v64/github"

	"github.com/tanush-128/metahost/models"
	"github.com/tanush-128/metahost/service"
	"github.com/tanush-128/metahost/utils"
)

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
				appService.RunApp(appConf)
			}

		} else if *latestCommit.SHA != *commits[0].SHA {
			latestCommit = commits[0]

			log.Printf("new found commit : %s \n", *latestCommit.Commit.Message)

			appConfs, err := GetAppConfs(client, repoConfig)
			if err != nil {
				fmt.Println(err)
			}

			for _, appConf := range appConfs {
				appService.RunApp(appConf)
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
