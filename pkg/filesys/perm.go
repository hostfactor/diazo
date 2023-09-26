package filesys

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
)

type FilePerm uint32

const (
	FilePermRead FilePerm = 1 << (32 - 1 - iota)
	FilePermCreate
	FilePermEdit
	FilePermDelete
)

func Perm(perms ...filesystem.FileAccessPolicy_FilePerm) FilePerm {
	b := FilePerm(0)
	for _, v := range perms {
		if v == filesystem.FileAccessPolicy_none {
			return 0
		}
		b |= ToFilePerm(v)
	}
	return b
}

func ToFilePerm(perm filesystem.FileAccessPolicy_FilePerm) FilePerm {
	o := FilePerm(0)
	switch perm {
	case filesystem.FileAccessPolicy_read:
		o = FilePermRead
	case filesystem.FileAccessPolicy_create:
		o = FilePermCreate
	case filesystem.FileAccessPolicy_edit:
		o = FilePermEdit
	case filesystem.FileAccessPolicy_delete:
		o = FilePermDelete
	}
	return o
}
