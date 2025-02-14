package plugin

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/drone/drone-go/drone"

	"github.com/bitsbeats/drone-tree-config/plugin/scm_clients"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// NewScmClient creates a new client for the git provider
func (p *Plugin) NewScmClient(ctx context.Context, uuid uuid.UUID, repo drone.Repo) (scmClient scm_clients.ScmClient, err error) {
	switch {
	case p.gitHubToken != "":
		scmClient, err = scm_clients.NewGitHubClient(ctx, uuid, p.server, p.gitHubToken, repo)
	case p.gitLabToken != "":
		scmClient, err = scm_clients.NewGitLabClient(ctx, uuid, p.gitLabServer, p.gitLabToken, repo)
	case p.bitBucketClient != "":
		scmClient, err = scm_clients.NewBitBucketClient(uuid, p.bitBucketAuthServer, p.server, p.bitBucketClient, p.bitBucketSecret, repo)
	case p.giteaToken != "":
		scmClient, err = scm_clients.NewGiteaClient(p.giteaServer, p.giteaToken, repo)
	default:
		err = fmt.Errorf("no SCM credentials specified")
	}
	if err != nil {
		return nil, fmt.Errorf("unable to connect to SCM server: %s", err)
	}
	return
}

// getChanges tries to get a list of changed files from github
func (p *Plugin) getScmChanges(ctx context.Context, req *request) ([]string, error) {
	var changedFiles []string

	if req.Build.Trigger == "@cron" {
		// cron jobs trigger a full build
		changedFiles = []string{}
	} else if strings.HasPrefix(req.Build.Ref, "refs/pull/") {
		// use pullrequests api to get changed files
		pullRequestID, err := strconv.Atoi(strings.Split(req.Build.Ref, "/")[2])
		if err != nil {
			logrus.Errorf("%s unable to get pull request id %v", req.UUID, err)
			return nil, err
		}
		changedFiles, err = req.Client.ChangedFilesInPullRequest(ctx, pullRequestID)
		if err != nil {
			logrus.Errorf("%s unable to fetch diff for Pull request %v", req.UUID, err)
		}
	} else if strings.HasPrefix(req.Build.Ref, "refs/tags/") {
		tagType := strings.Split(req.Build.Ref, "/")[2]
		// 发布生产环境 tag
		if strings.HasPrefix(tagType, "release") {
			changedFiles = p.allService
			return changedFiles, nil
		} else if strings.HasPrefix(tagType, "hotfix") {
			tagsSha, err := req.Client.GetTagShaList(ctx, "")
			if err != nil {
				return nil, err
			}
			logrus.Infof("tagsSha: %+v", tagsSha[:2])

			changedFiles, err = req.Client.ChangedFilesInDiff(ctx, tagsSha[1], tagsSha[0])
			if err != nil {
				logrus.Errorf("%s unable to fetch diff: '%v'", req.UUID, err)
				return nil, err
			}
		} else if strings.HasPrefix(tagType, "v") {
			tagsSha, err := req.Client.GetTagShaList(ctx, "v")
			if err != nil {
				return nil, err
			}
			logrus.Infof("tagsSha: %+v", tagsSha[:2])

			changedFiles, err = req.Client.ChangedFilesInDiff(ctx, tagsSha[1], tagsSha[0])
			if err != nil {
				logrus.Errorf("%s unable to fetch diff: '%v'", req.UUID, err)
				return nil, err
			}
		}
	} else {
		// use diff to get changed files
		before := req.Build.Before
		after := req.Build.After

		// check for broken before
		if before == "0000000000000000000000000000000000000000" || before == "" {
			before = fmt.Sprintf("%s~1", after)
		}

		var err error
		changedFiles, err = req.Client.ChangedFilesInDiff(ctx, before, after)
		if err != nil {
			logrus.Errorf("%s unable to fetch diff: '%v'", req.UUID, err)
			return nil, err
		}
	}

	if len(changedFiles) > 0 {
		logrus.Infof("%s changed files: %+v", req.UUID, changedFiles)
	} else {
		return nil, nil
	}
	return changedFiles, nil
}

// getFile downloads a file from github
func (p *Plugin) getScmFile(ctx context.Context, req *request, file string) (content string, err error) {
	logrus.Debugf("%s checking %s/%s %s", req.UUID, req.Repo.Namespace, req.Repo.Name, file)
	return req.Client.GetFileContents(ctx, file, req.Build.After)
}
