package units

import (
	"context"
	"testing"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/cloud"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/testutil"
	"github.com/evergreen-ci/utility"
	"github.com/stretchr/testify/assert"
)

func TestSpawnhostStartJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = testutil.TestSpan(ctx, t)

	assert.NoError(t, db.ClearCollections(host.Collection, event.EventCollection))
	mock := cloud.GetMockProvider()
	t.Run("NewSpawnhostStartJobHostNotStopped", func(t *testing.T) {
		tctx := testutil.TestSpan(ctx, t)
		h := host.Host{
			Id:       "host-running",
			Status:   evergreen.HostRunning,
			Provider: evergreen.ProviderNameMock,
			Distro:   distro.Distro{Provider: evergreen.ProviderNameMock},
		}
		assert.NoError(t, h.Insert(tctx))
		mock.Set(h.Id, cloud.MockInstance{
			Status: cloud.StatusRunning,
		})

		ts := utility.RoundPartOfMinute(1).Format(TSFormat)
		j := NewSpawnhostStartJob(&h, "user", ts)

		j.Run(context.Background())
		assert.Error(t, j.Error())

		checkSpawnHostModificationEvent(t, h.Id, event.EventHostStarted, false)
	})
	t.Run("NewSpawnhostStartJobOK", func(t *testing.T) {
		tctx := testutil.TestSpan(ctx, t)
		h := host.Host{
			Id:       "host-stopped",
			Status:   evergreen.HostStopped,
			Provider: evergreen.ProviderNameMock,
			Distro:   distro.Distro{Provider: evergreen.ProviderNameMock},
		}
		assert.NoError(t, h.Insert(tctx))
		mock.Set(h.Id, cloud.MockInstance{
			Status: cloud.StatusStopped,
		})

		ts := utility.RoundPartOfMinute(1).Format(TSFormat)
		j := NewSpawnhostStartJob(&h, "user", ts)

		j.Run(context.Background())
		assert.NoError(t, j.Error())
		assert.True(t, j.Status().Completed)

		startedHost, err := host.FindOneId(tctx, h.Id)
		assert.NoError(t, err)
		assert.Equal(t, evergreen.HostRunning, startedHost.Status)

		checkSpawnHostModificationEvent(t, h.Id, event.EventHostStarted, true)
	})
}
