package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const (
	owner        = "opencredo"
	repo         = "terrahelp"
	releaseTag   = "v0.7.5"
	assetName    = "terrahelp_0.7.5_linux_amd64.tar.gz"
	binaryFolder = "bin"
)

func main() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN not set.")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	assetURL, err := getReleaseAssetURL(ctx, client, owner, repo, releaseTag, assetName)
	if err != nil {
		fmt.Printf("Error getting asset URL: %s\n", err)
		os.Exit(1)
	}

	binaryPath, err := downloadReleaseAsset(ctx, client, assetURL, binaryFolder, assetName)
	if err != nil {
		fmt.Printf("Error downloading release asset: %s\n", err)
		os.Exit(1)
	}

	err = addBinaryToPath(binaryPath)
	if err != nil {
		fmt.Printf("Error setting PATH: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s is now available on PATH.\n", assetName)
}

func getReleaseAssetURL(ctx context.Context, client *github.Client, owner, repo, releaseTag, assetName string) (string, error) {
	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, releaseTag)
	if err != nil {
		return "", err
	}

	for _, asset := range release.Assets {
		if *asset.Name == assetName {
			return *asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("asset '%s' not found in release '%s'", assetName, releaseTag)
}

func downloadReleaseAsset(ctx context.Context, client *github.Client, assetURL, destFolder, assetName string) (string, error) {
	resp, err := client.Client().Get(assetURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download asset: %s", resp.Status)
	}

	binaryPath := filepath.Join(destFolder, assetName)
	out, err := os.Create(binaryPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return "", nil
}

func addBinaryToPath(binaryPath string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("setting PATH is not supported on Windows yet")
	}

	// Add the binary folder to PATH
	err := os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), binaryFolder))
	if err != nil {
		return err
	}

	return nil
}
