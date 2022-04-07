package containers

import (
	"context"
	"github.com/containers/common/pkg/auth"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"regexp"
)

var DefaultClient, _ = NewClient()

// Architecture specifies the architecture to fetch for the container. See https://docs.docker.com/desktop/multi-arch/ for more info and options.
var Architecture = "amd64"

// Platform specifies the platform to fetch for the container. See https://docs.docker.com/desktop/multi-arch/ for more info and options.
var Platform = "linux"

type Client interface {
	GetRepositoryTags(i ImageURL, opts GetRepositoryTagsOpts) ([]string, error)
	Login(username, password, registry string) error
	CopyImage(src, dst ImageURL) error
}

func NewClient() (Client, error) {
	policyCtx, err := defaultPolicyContext()
	if err != nil {
		return nil, err
	}

	return &client{
		SysCtx:    defaultSysCtx(),
		PolicyCtx: policyCtx,
	}, nil
}

// GetRepositoryTags Lists all tags for an image from a remote repository.
func GetRepositoryTags(i ImageURL, opts GetRepositoryTagsOpts) ([]string, error) {
	return DefaultClient.GetRepositoryTags(i, opts)
}

// Login logs into a remote registry and adds the credentials to the client.
func Login(username, password, registry string) error {
	return DefaultClient.Login(username, password, registry)
}

// CopyImage copies an image from a remote repository to another.
func CopyImage(src, dst ImageURL) error {
	return DefaultClient.CopyImage(src, dst)
}

type GetRepositoryTagsOpts struct {
	// Filter all tags with a regex. If the regex has a match, it will be included.
	Regexp *regexp.Regexp
}

type client struct {
	SysCtx    *types.SystemContext
	PolicyCtx *signature.PolicyContext
}

func (c *client) GetRepositoryTags(i ImageURL, opts GetRepositoryTagsOpts) ([]string, error) {
	ref, err := parseImageRef(i)
	if err != nil {
		return nil, err
	}

	tags, err := docker.GetRepositoryTags(context.Background(), defaultSysCtx(), ref)
	if err != nil {
		return nil, err
	}

	if opts.Regexp != nil {
		filteredTags := make([]string, 0, len(tags))
		for _, v := range tags {
			if opts.Regexp.MatchString(v) {
				filteredTags = append(filteredTags, v)
			}
		}
		tags = filteredTags
	}

	return tags, nil
}

func (c *client) Login(username, password, registry string) error {
	opts := &auth.LoginOptions{
		Password: password,
		Username: username,
		Stdout:   os.Stdout,
	}

	err := auth.Login(context.Background(), defaultSysCtx(), opts, []string{registry})
	if err != nil {
		return err
	}

	return nil
}

func ListTags(imageUrl ImageURL) ([]string, error) {
	ref, err := parseImageRef(imageUrl)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	tags, err := docker.GetRepositoryTags(ctx, defaultSysCtx(), ref)
	if err != nil {
		logrus.WithError(err).WithField("ref", imageUrl).Error("Failed to list tags.")
		return nil, err
	}

	return tags, nil
}

func (c *client) CopyImage(src, dst ImageURL) error {
	srcRef, err := parseImageRef(src)
	if err != nil {
		return err
	}

	dstRef, err := parseImageRef(dst)
	if err != nil {
		return err
	}

	policyCtx, err := defaultPolicyContext()
	if err != nil {
		return err
	}

	sysCtx := defaultSysCtx()

	ctx := context.Background()

	opts := &copy.Options{
		ReportWriter:   io.Discard,
		SourceCtx:      sysCtx,
		DestinationCtx: sysCtx,
	}

	_, err = copy.Image(ctx, policyCtx, dstRef, srcRef, opts)
	if err != nil {
		return err
	}

	return nil
}

func parseImageRef(i ImageURL) (types.ImageReference, error) {
	ref, err := alltransports.ParseImageName(i.String())
	if err != nil {
		return nil, err
	}

	return ref, nil
}

func defaultPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}

	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		return nil, err
	}

	return policyCtx, nil
}

func defaultSysCtx() *types.SystemContext {
	return &types.SystemContext{
		ArchitectureChoice: Architecture,
		OSChoice:           Platform,
	}
}
