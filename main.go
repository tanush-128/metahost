package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

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

	file, dirs, _, _ := client.Repositories.GetContents(ctx, "tanush-128", "metaserver-conf", "", nil)

	if file != nil {

		fmt.Print(file.Content)
	}
	for _, dir := range dirs {
		file, _, _, _ := client.Repositories.GetContents(ctx, "tanush-128", "metaserver-conf", dir.GetPath(), nil)
		fmt.Print(file.GetContent())
	}

	var latestCommit *github.RepositoryCommit

	for {
		commits, _, err := client.Repositories.ListCommits(ctx, "tanush-128", "metaserver-conf", nil)
		if err != nil {
			log.Println(err)
		}
		if latestCommit == nil {
			latestCommit = commits[0]
			fmt.Printf("last found commit : %s \n", *latestCommit.Commit.Message)
		} else if *latestCommit.SHA != *commits[0].SHA {
			latestCommit = commits[0]
			log.Printf("new found commit : %s \n", *latestCommit.Commit.Message)
		} else {

			log.Println("No New Commit Found")
		}

		time.Sleep(10 * time.Second)
	}

	// repo, _, err := client.Repositories.Get(ctx, "tanush-128", "metaserver-conf")
	// if err != nil {
	// 	log.Println(err)

	// }

}
