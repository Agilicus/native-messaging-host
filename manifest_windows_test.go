//go:build windows

package host

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestManifestTargetName(t *testing.T) {
	t.Parallel()

	got, err := (&Host{AppName: "app"}).getTargetNames()
	assert.Nil(t, err)
	assert.Greater(t, len(got), 0)
	want := `SOFTWARE\Google\Chrome`

	if diff := cmp.Diff(want, got[0]); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
