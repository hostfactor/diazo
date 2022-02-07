package containers

import (
	"context"
	"github.com/containers/common/pkg/auth"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"io"
	"os"
)

var DefaultClient Client

type Client interface {
	GetRepositoryTags(i ImageURL) ([]string, error)
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

type client struct {
	SysCtx    *types.SystemContext
	PolicyCtx *signature.PolicyContext
}

func (c *client) GetRepositoryTags(i ImageURL) ([]string, error) {
	ref, err := parseImageRef(i)
	if err != nil {
		return nil, err
	}

	tags, err := docker.GetRepositoryTags(context.Background(), defaultSysCtx(), ref)
	if err != nil {
		return nil, err
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
		ArchitectureChoice: "amd64",
		OSChoice:           "linux",
	}
}
