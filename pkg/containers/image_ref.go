package containers

import "fmt"

type ImageURL interface {
	fmt.Stringer
	getProto() string
}

type DockerImageUrl string

func (d DockerImageUrl) String() string { return toString(d, string(d)) }

func (d DockerImageUrl) getProto() string { return "docker" }

func toString(i ImageURL, url string) string {
	return i.getProto() + "://" + url
}
