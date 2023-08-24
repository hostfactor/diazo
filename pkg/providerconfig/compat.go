package providerconfig

import (
	"github.com/hostfactor/api/go/blueprint/steps"
	"github.com/hostfactor/api/go/providerconfig"
	"github.com/hostfactor/diazo/pkg/collection"
	"github.com/hostfactor/diazo/pkg/ptr"
)

// The old style of form that applies to every app.
func defaultForms(conf *providerconfig.ProviderConfig) []*providerconfig.SettingsForm {
	out := []*providerconfig.SettingsForm{
		{
			Steps: []*steps.Step{
				{
					Id:    "settings",
					Title: ptr.Ptr("Settings"),
					Components: []*steps.Component{
						{
							Id:      "version",
							Version: &steps.VersionComponent{},
						},
					},
				},
			},
		},
	}

	out[0].Steps[0].Components = append(out[0].Steps[0].Components, collection.Map(conf.GetVolumes(), func(f *providerconfig.Volume) *steps.Component {
		fi := f.GetSource().GetFileInput()
		return &steps.Component{
			Id:    f.Name,
			Style: &steps.ComponentStyle{Width: ptr.Ptr[int32](8)},
			FileSelect: &steps.FileSelectComponent{
				AcceptProps: fi.GetAcceptProps(),
				HelpText:    ptr.NonZeroPtr(fi.GetHelpText()),
				Disabled:    ptr.NonZeroPtr(fi.GetDisabled()),
				Multiple:    ptr.NonZeroPtr(fi.GetMultiple()),
				Volume:      f.Name,
				Description: ptr.NonZeroPtr(fi.GetDescription()),
				ZipFilename: ptr.NonZeroPtr(fi.GetZipFilename()),
				Title:       ptr.NonZeroPtr(fi.GetTitle()),
				Matcher:     fi.GetMatcher(),
				Folder:      ptr.NonZeroPtr(fi.GetDestination().GetBucketFolder()),
			},
		}
	})...)

	out[0].Steps[0].Components = append(out[0].Steps[0].Components, &steps.Component{
		Id: "settings",
		JsonSchema: &steps.JSONSchemaComponent{
			Path: ptr.String("settings.json"),
		},
		Title: ptr.Ptr("Settings"),
	})

	return out
}
