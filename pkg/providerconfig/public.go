package providerconfig

import (
	"github.com/hostfactor/api/go/blueprint"
	"github.com/hostfactor/api/go/blueprint/actions"
	"github.com/hostfactor/api/go/blueprint/filesystem"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/volume"
)

func ToSetupActions(data *blueprint.BlueprintData, p *providerconfig.ProviderConfig) []*blueprint.SetupAction {
	acts := make([]*blueprint.SetupAction, 0, len(data.GetSelectedFiles()))
	volMap := volume.VolumesToMap(p.GetVolumes())
	for _, selection := range data.GetSelectedFiles() {
		vol := volMap[selection.GetVolumeName()]
		if vol == nil {
			continue
		}
		sa := MountFileSelectionSetupAction(selection, vol.GetMount())
		if sa != nil {
			acts = append(acts, sa)
		}
	}

	return acts
}

func MountFileSelectionSetupAction(sel *filesystem.FileSelection, mount *providerconfig.VolumeMount) *blueprint.SetupAction {
	for _, loc := range sel.GetLocations() {
		if source := loc.GetBucketFile(); source != nil {
			if source.GetName() == "" {
				continue
			}
			return &blueprint.SetupAction{

				Download: &actions.DownloadFile{
					Source: &actions.DownloadFile_Source{Storage: &filesystem.BucketFileMatcher{
						Matches: &filesystem.FileMatcher{Name: source.GetName()},
						Folder:  source.GetFolder(),
					}},
					To: mount.GetPath(),
				},
			}
		}
	}
	return nil
}
