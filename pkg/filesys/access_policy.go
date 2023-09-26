package filesys

import "github.com/hostfactor/api/go/blueprint/filesystem"

func NewAccessPolicy(ap *filesystem.FileAccessPolicy) *AccessPolicy {
	return &AccessPolicy{
		Wrapped: ap,
		Perm:    Perm(ap.GetPerms()...),
	}
}

type AccessPolicy struct {
	Wrapped *filesystem.FileAccessPolicy
	Perm    FilePerm
}
