package scm_clients

import (
	"context"
	"time"

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

type result struct {
	ChangedFiles []string
	Err          error
}

func (s GiteaClient) ChangedFilesInDiff(ctx context.Context, base string, head string) ([]string, error) {
	logrus.Infof("ChangedFilesInDiff, base:%s, head:%s", base, head)
	nowCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, head)
	if err != nil {
		logrus.Error("GetSingleCommit, err:", err)
		return nil, err
	}

	resultChan := make(chan result)
	go func() {
		currentCommit := nowCommit
		var changedFiles []string
		for {
			if len(currentCommit.Parents) == 1 {
				if currentCommit.SHA == base {
					break
				} else {
					for _, file := range currentCommit.Files {
						changedFiles = append(changedFiles, file.Filename)
					}
					parentCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, currentCommit.Parents[0].SHA)
					if err != nil {
						logrus.Errorf("GetSingleCommit, err: %v", err)
					}
					currentCommit = parentCommit
				}
			} else {
				logrus.Infof("commit.Parents len is : %d, not implemente", len(currentCommit.Parents))
				break
			}
		}
		resultChan <- result{
			ChangedFiles: changedFiles,
			Err:          err,
		}
	}()

	// resultChan := make(chan result)
	//
	// go func() {
	// 	var changedFiles []string
	// 	var err error
	// find:
	// 	for {
	// 		newCurrentCommits := make([]*gitea.Commit, 0)
	// 		for _, commit := range currentCommits {
	// 			if commit.SHA == base {
	// 				break find
	// 			} else {
	// 				for _, file := range commit.Files {
	// 					// logrus.Infof("parentCommit file.Filename: %s", file.Filename)
	// 					changedFiles = append(changedFiles, file.Filename)
	// 				}
	// 			}
	//
	// 			switch len(commit.Parents) {
	// 			case 1:
	// 				parentCommit, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, commit.Parents[0].SHA)
	// 				if err != nil {
	// 					logrus.Error("GetSingleCommit, err:", err)
	// 				}
	// 				newCurrentCommits = append(newCurrentCommits, parentCommit)
	// 			// case 2:
	// 			// 	logrus.Infof("commit.Parents len is : %d\n", len(commit.Parents))
	// 			// 	logrus.Infof("commit.Parents.SHA: %s\n", commit.Parents[0].SHA)
	// 			// 	logrus.Infof("commit.Parents.SHA: %s\n", commit.Parents[1].SHA)
	// 			//
	// 			// 	parentCommit1, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, commit.Parents[0].SHA)
	// 			// 	if err != nil {
	// 			// 		logrus.Error("GetSingleCommit, err:", err)
	// 			// 		return nil, err
	// 			// 	}
	// 			// 	newCurrentCommits = append(newCurrentCommits, parentCommit1)
	// 			//
	// 			// 	parentCommit2, _, err := s.delegate.GetSingleCommit(s.repo.Namespace, s.repo.Name, commit.Parents[1].SHA)
	// 			// 	if err != nil {
	// 			// 		logrus.Error("GetSingleCommit, err:", err)
	// 			// 		return nil, err
	// 			// 	}
	// 			// 	newCurrentCommits = append(newCurrentCommits, parentCommit2)
	//
	// 			// 多个Parents，不处理，直接退出
	// 			default:
	// 				logrus.Infof("commit.Parents len is : %d, not implemente", len(commit.Parents))
	// 				break find
	// 			}
	// 		}
	// 		currentCommits = newCurrentCommits
	// 	}
	// 	resultChan <- result{
	// 		ChangedFiles: changedFiles,
	// 		Err:          err,
	// 	}
	// }()

	var changedFiles []string
	select {
	case r := <-resultChan:
		if r.Err != nil {
			logrus.Errorf("get changedFiles error: %v", r.Err)
			return nil, r.Err
		} else {
			changedFiles = r.ChangedFiles
			// 去重
			outputChangedFiles := make([]string, 0)
			seen := map[string]bool{}
			for _, value := range changedFiles {
				if seen[value] == true {
					continue
				}
				seen[value] = true
				outputChangedFiles = append(outputChangedFiles, value)
			}

			// logrus.Infof("changedFiles: %+v", outputChangedFiles)
			return outputChangedFiles, nil
		}
	// 加个超时控制
	case <-time.After(10 * time.Second):
		logrus.Infof("get changedFiles timeout")
		return nil, nil
	}
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
