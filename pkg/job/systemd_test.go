package job

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateExecStart(t *testing.T) {
	tc := []struct {
		givenRootPath string
		argv          []string
		wd            string
		execStart     string
	}{
		{
			"/opt/sandbox",
			[]string{"/opt/bin/pupernetes", "run", "/opt/sandbox", "--job-type", "systemd"},
			"/",
			"/opt/bin/pupernetes run /opt/sandbox --job-type fg",
		},
		{
			"/opt/sandbox",
			[]string{"/opt/bin/pupernetes", "run", "/opt/sandbox", "--job-type=systemd"},
			"/",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg",
		},
		{
			"/opt/sandbox",
			[]string{"/opt/bin/pupernetes", "run", "/opt/sandbox", "--job-type=systemd", "-v", "3"},
			"/",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg -v 3",
		},
		{
			"sandbox",
			[]string{"/opt/bin/pupernetes", "run", "sandbox", "--job-type=systemd", "-v", "3"},
			"/opt",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg -v 3",
		},
		{
			"./sandbox",
			[]string{"/opt/bin/pupernetes", "run", "./sandbox", "--job-type=systemd", "-v", "3"},
			"/opt",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg -v 3",
		},
		{
			"../sandbox",
			[]string{"/opt/bin/pupernetes", "run", "../sandbox", "--job-type=systemd", "-v", "3"},
			"/opt/bin",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg -v 3",
		},
		{
			"../sandbox",
			[]string{"/opt/bin/pupernetes", "run", "../sandbox", "--job-type=systemd", "-v", "3", "-c", "systemd"},
			"/opt/bin",
			"/opt/bin/pupernetes run /opt/sandbox --job-type=fg -v 3 -c systemd",
		},
		{
			"../sandbox",
			[]string{"/opt/bin/pupernetes", "run", "../sandbox", "-c", "systemd", "--job-type=systemd", "-v", "3"},
			"/opt/bin",
			"/opt/bin/pupernetes run /opt/sandbox -c systemd --job-type=fg -v 3",
		},
		{
			"../sandbox",
			[]string{"/opt/bin/pupernetes", "run", "../sandbox", "--clean=systemd", "--job-type=systemd", "-v", "3"},
			"/opt/bin",
			"/opt/bin/pupernetes run /opt/sandbox --clean=systemd --job-type=fg -v 3",
		},
	}
	for _, elt := range tc {
		t.Run("", func(t *testing.T) {
			s, err := createExecStart(elt.givenRootPath, elt.argv, elt.wd)
			require.NoError(t, err)
			assert.Equal(t, elt.execStart, s)
		})
	}
}
