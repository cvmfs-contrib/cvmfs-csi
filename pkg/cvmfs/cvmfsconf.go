package cvmfs

import (
	"io/ioutil"
	"os"
	"path"
	"text/template"
)

const (
	cvmfsConfigRoot = "/etc/cvmfs"
)

const repoConf = `
{{fileContents "/etc/cvmfs/default.conf"}}
{{fileContents "/etc/cvmfs/default.local"}}

{{if .Proxy}}
CVMFS_HTTP_PROXY={{.Proxy}}
{{end}}

CVMFS_CACHE_BASE={{cacheBase .VolumeId}}

{{if .Hash}}
CVMFS_ROOT_HASH={{.Hash}}
CVMFS_AUTO_UPDATE=no
{{else if .Tag}}
CVFMFS_REPOSITORY_TAG={{.Tag}}
{{end}}`

var (
	repoConfTempl *template.Template
)

func init() {
	fs := map[string]interface{}{
		"fileContents": func(filePath string) string {
			if c, err := ioutil.ReadFile(filePath); err != nil {
				panic(err)
			} else {
				return string(c)
			}
		},
		"cacheBase": getVolumeCachePath,
	}

	repoConfTempl = template.Must(template.New("repo-conf").Funcs(fs).Parse(repoConf))
}

type cvmfsConfigData struct {
	VolumeId  volumeID
	Tag, Hash string
	Proxy     string
}

func (d *cvmfsConfigData) writeToFile() error {
	f, err := os.OpenFile(getConfigFilePath(d.VolumeId), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0755)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}

	defer f.Close()

	return repoConfTempl.Execute(f, d)
}

func getConfigFilePath(volId volumeID) string {
	return path.Join(cvmfsConfigRoot, "config-csi-"+string(volId))
}
