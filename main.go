package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

const (
	owner        = "opencredo"
	repo         = "terrahelp"
	releaseTag   = "v0.7.5"
	assetName    = "terrahelp"
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

	tag := strings.ReplaceAll(releaseTag, "v", "")
	releaseAssetName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", assetName, tag, runtime.GOOS, runtime.GOARCH)

	client := github.NewClient(tc)
	assetURL, err := getReleaseAssetURL(ctx, client, owner, repo, releaseTag, releaseAssetName)
	if err != nil {
		fmt.Printf("Error getting asset URL: %s\n", err)
		os.Exit(1)
	}

	binaryPath, err := downloadReleaseAsset(ctx, client, assetURL, binaryFolder, releaseAssetName)
	if err != nil {
		fmt.Printf("Error downloading release asset: %s\n", err)
		os.Exit(1)
	}

	err = unpackTarGz(binaryPath)
	if err != nil {
		fmt.Printf("Error unpacking release asset: %s\n", err)
		os.Exit(1)
	}

	err = addBinaryToPath(client, binaryFolder)
	if err != nil {
		fmt.Printf("Error setting PATH: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s is now available on PATH\n", releaseAssetName)
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

	err = os.MkdirAll(destFolder, os.ModePerm)
	if err != nil {
		return "", err
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

	return binaryPath, nil
}

func unpackTarGz(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %s", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %s", err)
		}

		destPath := filepath.Join(filepath.Dir(filePath), header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %s", err)
			}
		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create parent directory: %s", err)
			}

			// Create the file
			unzippedFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %s", err)
			}
			defer unzippedFile.Close()

			// Copy contents from tar file to the destination file
			if _, err := io.Copy(unzippedFile, tr); err != nil {
				return fmt.Errorf("failed to copy file contents: %s", err)
			}
		}
	}

	return nil
}

func addBinaryToPath(client *github.Client, binaryFolder string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("setting PATH is not supported on Windows yet")
	}

	// Add the binary folder to PATH
	cmd := exec.Command("echo", fmt.Sprintf("$HOME/%s >> $GITHUB_PATH", binaryFolder))
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}
