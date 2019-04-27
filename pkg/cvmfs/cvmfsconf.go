package cvmfs

import (
	"io/ioutil"
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

