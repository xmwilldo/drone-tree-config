package main

import (
	"net/http"
	"time"

	"github.com/bitsbeats/drone-tree-config/plugin"

	"github.com/drone/drone-go/plugin/config"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type (
	spec struct {
		AllowListFile       string        `envconfig:"PLUGIN_ALLOW_LIST_FILE"`
		Concat              bool          `envconfig:"PLUGIN_CONCAT"`
		MaxDepth            int           `envconfig:"PLUGIN_MAXDEPTH" default:"2"`
		Fallback            bool          `envconfig:"PLUGIN_FALLBACK"`
		Debug               bool          `envconfig:"PLUGIN_DEBUG"`
		Address             string        `envconfig:"PLUGIN_ADDRESS" default:":3000"`
		Secret              string        `envconfig:"PLUGIN_SECRET"`
		Server              string        `envconfig:"SERVER" default:"https://api.github.com"`
		GitHubToken         string        `envconfig:"GITHUB_TOKEN"`
		GitLabToken         string        `envconfig:"GITLAB_TOKEN"`
		GitLabServer        string        `envconfig:"GITLAB_SERVER" default:"https://gitlab.com"`
		BitBucketAuthServer string        `envconfig:"BITBUCKET_AUTH_SERVER"`
		BitBucketClient     string        `envconfig:"BITBUCKET_CLIENT"`
		BitBucketSecret     string        `envconfig:"BITBUCKET_SECRET"`
		GiteaServer         string        `envconfig:"GITEA_SERVER"`
		GiteaToken          string        `envconfig:"GITEA_TOKEN"`
		ConsiderFile        string        `envconfig:"PLUGIN_CONSIDER_FILE"`
		CacheTTL            time.Duration `envconfig:"PLUGIN_CACHE_TTL"`
		AllService          string        `envconfig:"ALL_SERVICE"`
	}
)

func main() {
	spec := new(spec)
	if err := envconfig.Process("", spec); err != nil {
		logrus.Fatal(err)
	}

	if spec.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if spec.Secret == "" {
		logrus.Fatalln("missing secret key")
	}
	if spec.GitHubToken == "" && spec.GitLabToken == "" && spec.GiteaToken == "" && (spec.BitBucketClient == "" || spec.BitBucketSecret == "") {
		logrus.Warnln("missing SCM credentials, e.g. GitHub token")
	}
	if spec.Address == "" {
		spec.Address = ":3000"
	}
	if spec.BitBucketAuthServer == "" {
		spec.BitBucketAuthServer = spec.Server
	}

	handler := config.Handler(
		plugin.New(
			plugin.WithConcat(spec.Concat),
			plugin.WithFallback(spec.Fallback),
			plugin.WithMaxDepth(spec.MaxDepth),
			plugin.WithServer(spec.Server),
			plugin.WithAllowListFile(spec.AllowListFile),
			plugin.WithBitBucketAuthServer(spec.BitBucketAuthServer),
			plugin.WithBitBucketClient(spec.BitBucketClient),
			plugin.WithBitBucketSecret(spec.BitBucketSecret),
			plugin.WithGithubToken(spec.GitHubToken),
			plugin.WithGitlabToken(spec.GitLabToken),
			plugin.WithGitlabServer(spec.GitLabServer),
			plugin.WithGiteaServer(spec.GiteaServer),
			plugin.WithGiteaToken(spec.GiteaToken),
			plugin.WithConsiderFile(spec.ConsiderFile),
			plugin.WithCacheTTL(spec.CacheTTL),
			plugin.WithAllService(spec.AllService),
		),
		spec.Secret,
		logrus.StandardLogger(),
	)

	logrus.Infof("server listening on address %s", spec.Address)

	http.Handle("/", handler)
	logrus.Fatal(http.ListenAndServe(spec.Address, nil))
}
