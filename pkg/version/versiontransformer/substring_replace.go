package versiontransformer

import (
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/version/remoteversion"
	"regexp"
)

func NewSubstringReplace(s *providerconfig.SubstringReplace) (Transformer, error) {
	r, err := regexp.Compile(s.Matches)
	if err != nil {
		return nil, err
	}

	return &substringReplace{
		Repl:   s,
		Regexp: r,
	}, nil
}

type substringReplace struct {
	Repl   *providerconfig.SubstringReplace
	Regexp *regexp.Regexp
}

func (s *substringReplace) TransformRemoteVersion(ver *remoteversion.RemoteVersion) *remoteversion.RemoteVersion {
	name := s.Regexp.ReplaceAllString(ver.Name, s.Repl.Replacement)
	return &remoteversion.RemoteVersion{
		Name: name,
	}
}
