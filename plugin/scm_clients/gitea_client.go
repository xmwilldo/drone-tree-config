package scm_clients

import (
	"context"

	"code.gitea.io/sdk/gitea"
	"github.com/drone/drone-go/drone"
	"github.com/sirupsen/logrus"
)

type GiteaClient struct {
	delegate *gitea.Client
	repo     drone.Repo
}

func NewGiteaClient(server string, token string, repo drone.Repo) (ScmClient, error) {
	var client *gitea.Client
	var err error
	client, err = gitea.NewClient(server,
		gitea.SetToken(token),
		gitea.SetDebugMode())

	return GiteaClient{
		delegate: client,
		repo:     repo,
	}, err
}

func (s GiteaClient) ChangedFilesInPullRequest(ctx context.Context, pullRequestID int) ([]string, error) {
	return []string{}, nil
}

func (s GiteaClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {
	var changedFiles []string
	commit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, head)
	if err != nil {
		logrus.Error("GetSingleCommit, err:", err)
		return nil, err
	}

	for _, file := range commit.Files {
		changedFiles = append(changedFiles, file.Filename)
	}

	return changedFiles, nil
}

func (s GiteaClient) GetFileContents(ctx context.Context, path string, commitRef string) (string, error) {
	file, _, err := s.delegate.GetFile(s.repo.Namespace, s.repo.Name, commitRef, path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func (s GiteaClient) GetFileListing(ctx context.Context, path string, commitRef string) ([]FileListingEntry, error) {
	contents, _, err := s.delegate.ListContents(s.repo.Namespace, s.repo.Name, commitRef, path)
	if err != nil {
		return nil, err
	}

	var result []FileListingEntry
	for _, content := range contents {
		if content.Type != "file" && content.Type != "dir" {
			continue
		}
		fileListingEntry := FileListingEntry{
			Path: content.Path,
			Name: content.Name,
			Type: content.Type,
		}

		result = append(result, fileListingEntry)
	}

	return result, nil
}
