package containers

import (
	"errors"
	"path"
	"strings"
)

type ImageURLGetter interface {
	// String returns the full URL e.g. docker://docker.io/nginx.
	String() string

	// Path returns the Docker-compliant image path e.g. for "docker://docker.io/factoriotools/factorio:1.1",
	// "factoriotools/factorio:1.1" would be returned. Essentially the Repo + Name + [:]tag
	Path() string
}

type ImageURL struct {
	// Host is the base of the image URL e.g. for gcr.io/hostfactor/factorio:1.1 would be gcr.io. Defaults to docker.io.
	Host string

	// Repo is the portion of the image after the Host e.g. for gcr.io/hostfactor/factorio:1.1 would be hostfactor.
	// Most image URLs will have this but some Docker images e.g. "nginx" will not.
	Repo string

	// Name is the final portion of the slash-separated Path. e.g. for gcr.io/hostfactor/factorio:1.1 would be factorio.
	// All valid images will contain a Name.
	Name string

	// Tag is the value after the ":" character. e.g. for gcr.io/hostfactor/factorio:1.1 would be 1.1. defaults to "latest".
	Tag string

	// Proto is the protocol for this image. Defaults to "docker".
	Proto string
}

// Path returns the Docker-compliant image path e.g. for "docker://docker.io/factoriotools/factorio:1.1",
// "factoriotools/factorio:1.1" would be returned. Essentially the Repo + Name + [:]tag
func (i ImageURL) Path() string {
	injectDefaults(&i)
	builder := strings.Builder{}
	if i.Repo != "" {
		builder.WriteString(i.Repo + "/")
	}
	builder.WriteString(i.Name + ":" + i.Tag)
	return builder.String()
}

// String returns the full URL e.g. docker://docker.io/nginx.
func (i ImageURL) String() string {
	injectDefaults(&i)
	builder := strings.Builder{}
	builder.WriteString(i.Proto + "://" + i.Host + "/")
	if i.Repo != "" {
		builder.WriteString(i.Repo + "/")
	}
	builder.WriteString(i.Name + ":" + i.Tag)
	return builder.String()
}

func injectDefaults(i *ImageURL) {
	if i.Proto == "" {
		i.Proto = "docker"
	}

	if i.Host == "" {
		i.Host = "docker.io"
	}

	if i.Tag == "" {
		i.Tag = "latest"
	}
}

// ParseImageURL creates a docker URL from the form of [repository name]/image name. The repository name may be optional
// for certain images e.g. "nginx". Most images require the full URL e.g. "lloesche/valheim-server".
func ParseImageURL(img string) (ImageURL, error) {
	return parseImage(img)
}

func parseImage(img string) (ImageURL, error) {
	i := ImageURL{}
	injectDefaults(&i)
	schemeDelimiter := strings.Split(img, "://")
	pa := img
	switch len(schemeDelimiter) {
	case 1:
	case 2:
		i.Proto = schemeDelimiter[0]
		pa = schemeDelimiter[1]
	default:
		return i, errors.New("invalid url")
	}
	colonSeparated := strings.Split(pa, ":")
	switch len(colonSeparated) {
	case 1:
		i.Name = colonSeparated[0]
	case 2:
		pa = colonSeparated[0]
		i.Tag = colonSeparated[1]
	default:
		return i, errors.New("invalid url")
	}

	pathSplit := strings.Split(pa, "/")

	n := len(pathSplit)
	switch n {
	case 1:
		i.Name = pathSplit[0]
	case 2:
		if strings.Contains(pathSplit[0], ".") {
			i.Host = pathSplit[0]
		} else {
			i.Repo = pathSplit[0]
		}
		i.Name = pathSplit[1]
	default:
		i.Repo, i.Name = path.Split(strings.Join(pathSplit[1:], "/"))
		i.Repo = path.Clean(i.Repo)
		i.Host = pathSplit[0]
	}

	return i, nil
}
