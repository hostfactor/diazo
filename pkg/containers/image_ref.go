package containers

import "fmt"

type ImageURL interface {
	fmt.Stringer
	getProto() string
}

// DockerImageUrl creates a docker URL from the form of [repository name]/image name. The repository name may be optional
// for certain images e.g. "nginx". Most images require the full URL e.g. "lloesche/valheim-server".
type DockerImageUrl string

func (d DockerImageUrl) String() string { return toString(d, string(d)) }

func (d DockerImageUrl) getProto() string { return "docker" }

func toString(i ImageURL, url string) string {
	return i.getProto() + "://" + url
}
