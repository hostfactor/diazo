package fileselect

import (
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PublicTestSuite struct {
	suite.Suite
}

func (p *PublicTestSuite) TestMergeFileSelect() {
	// -- Given
	//
	type test struct {
		F1       []*filesystem.FileSelection
		F2       []*filesystem.FileSelection
		Expected func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection
	}

	tests := []test{
		{
			Expected: func(_, _ []*filesystem.FileSelection) []*filesystem.FileSelection {
				return []*filesystem.FileSelection{}
			},
		},
		{
			F1: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			Expected: func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
				return f1
			},
		},
		{
			F2: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			Expected: func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
				return f2
			},
		},
		{
			F1: []*filesystem.FileSelection{
				{
					VolumeName: "derp1",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			F2: []*filesystem.FileSelection{
				{
					VolumeName: "derp2",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			Expected: func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
				return append(f1, f2...)
			},
		},
		{
			F1: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.txt",
							Folder: "saves",
						}}},
					},
				},
			},
			F2: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			Expected: func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
				return f2
			},
		},
		{
			F1: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.txt",
							Folder: "saves",
						}}},
					},
				},
				{
					VolumeName: "derp1",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.txt",
							Folder: "saves",
						}}},
					},
				},
			},
			F2: []*filesystem.FileSelection{
				{
					VolumeName: "derp",
					Locations: []*filesystem.FileLocation{
						{Loc: &filesystem.FileLocation_BucketFile{BucketFile: &filesystem.BucketFile{
							Name:   "save.zip",
							Folder: "saves",
						}}},
					},
				},
			},
			Expected: func(f1, f2 []*filesystem.FileSelection) []*filesystem.FileSelection {
				return append(f2, f1[1])
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		actual := MergeFileSelect(v.F1, v.F2)
		p.Equal(v.Expected(v.F1, v.F2), actual, "test %d", i)
	}

	// -- Then
	//
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
