package nodepacker

import "testing"

func TestLoadManifests(t *testing.T) {
	mfs, err := loadManifests([]string{"/Users/uwe/work/src/github.com/sourcegraph/deploy-sourcegraph/base"})
	if err != nil {
		t.Error(err)
	}

	t.Log(mfs)
}

func TestParseResource(t *testing.T) {
	v := parseResourceString("100M", false)
	t.Log(v)

	v = parseResourceString("1G", false)
	t.Log(v)

	v = parseResourceString("2Gi", false)
	t.Log(v)

	v = parseResourceString("2", true)
	t.Log(v)

	v = parseResourceString("100m", true)
	t.Log(v)
}
