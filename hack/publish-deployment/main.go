package main

import (
	"context"
	"os"

	"github.com/google/go-github/v25/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()

	// Setup GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: os.Getenv("GITHUB_TOKEN"),
	})
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	// Setup logger
	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}

	// nolint: errcheck
	defer logger.Sync()

	// Read environment variables
	owner := os.Getenv("CIRCLE_PROJECT_USERNAME")
	repo := os.Getenv("CIRCLE_PROJECT_REPONAME")
	tag := os.Getenv("CIRCLE_TAG")

	// Get the release
	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)

	if err != nil {
		logger.Fatal("Failed to get the release", zap.Error(err))
	}

	logger.Info("Release", zap.Any("release", release))

	// Upload the asset
	asset, _, err := client.Repositories.UploadReleaseAsset(ctx, owner, repo, release.GetID(), &github.UploadOptions{
		Name: os.Getenv("ASSET_NAME"),
	}, os.Stdin)

	if err != nil {
		logger.Fatal("Failed to upload the asset", zap.Error(err))
	}

	logger.Info("Asset uploaded", zap.Any("asset", asset))
}