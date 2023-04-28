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
	logrus.Infof("ChangedFilesInDiff, base:%s, head:%s", base, head)
	var changedFiles []string
	commit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, head)
	if err != nil {
		logrus.Error("GetSingleCommit, err:", err)
		return nil, err
	}

	// files maybe null, find parent commit
	if len(commit.Files) == 0 {
		logrus.Info("commit.Files is empty, find parent commit")
		parentCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, commit.Parents[0].SHA)
		if err != nil {
			logrus.Error("GetSingleCommit, err:", err)
			return nil, err
		}
		thisCommit := parentCommit
		for _, file := range thisCommit.Files {
			logrus.Infof("conmmit file.Filename: %s", file.Filename)
			changedFiles = append(changedFiles, file.Filename)
		}

		for {
			if len(thisCommit.Parents) == 0 {
				break
			}

			logrus.Infof("commit.Parents.SHA: %s", thisCommit.Parents[0].SHA)
			parentCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, thisCommit.Parents[0].SHA)
			if err != nil {
				logrus.Error("GetSingleCommit, err:", err)
				return nil, err
			}
			thisCommit = parentCommit
			if thisCommit.SHA == base {
				break
			}
			for _, file := range thisCommit.Files {
				logrus.Infof("parentCommit file.Filename: %s", file.Filename)
				changedFiles = append(changedFiles, file.Filename)
			}
		}
	} else {
		logrus.Info("commit.Files is not empty")
		thisCommit := commit
		for _, file := range thisCommit.Files {
			logrus.Infof("conmmit file.Filename: %s", file.Filename)
			changedFiles = append(changedFiles, file.Filename)
		}

		for {
			if len(thisCommit.Parents) == 0 {
				break
			}

			logrus.Infof("commit.Parents.SHA: %s", thisCommit.Parents[0].SHA)
			parentCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, thisCommit.Parents[0].SHA)
			if err != nil {
				logrus.Error("GetSingleCommit, err:", err)
				return nil, err
			}
			thisCommit = parentCommit
			if thisCommit.SHA == base {
				break
			}
			for _, file := range thisCommit.Files {
				logrus.Infof("parentCommit file.Filename: %s", file.Filename)
				changedFiles = append(changedFiles, file.Filename)
			}
		}
	}

	logrus.Infof("changedFiles: %+v", changedFiles)
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
