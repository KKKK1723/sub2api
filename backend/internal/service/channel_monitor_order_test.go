//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type monitorOrderRepoStub struct {
	ChannelMonitorRepository
	calls int
	err   error
}

func (r *monitorOrderRepoStub) UpdateOrder(_ context.Context, _, _ []int64) error {
	r.calls++
	return r.err
}

func TestChannelMonitorOrderRejectsInvalidPermutation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		ids, expected []int64
	}{
		{"empty", nil, nil},
		{"missing", []int64{1}, []int64{1, 2}},
		{"duplicate", []int64{1, 1}, []int64{1, 2}},
		{"zero", []int64{0, 2}, []int64{0, 2}},
		{"negative", []int64{-1}, []int64{-1}},
		{"different", []int64{1, 3}, []int64{1, 2}},
		{"duplicate snapshot", []int64{1, 2}, []int64{1, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &monitorOrderRepoStub{}
			err := NewChannelMonitorService(repo, nil).UpdateOrder(context.Background(), tc.ids, tc.expected)
			require.Error(t, err)
			require.Zero(t, repo.calls)
		})
	}
}

func TestChannelMonitorOrderPropagatesConflict(t *testing.T) {
	repo := &monitorOrderRepoStub{err: ErrChannelMonitorOrderConflict}
	err := NewChannelMonitorService(repo, nil).UpdateOrder(context.Background(), []int64{2, 1}, []int64{1, 2})
	require.ErrorIs(t, err, ErrChannelMonitorOrderConflict)
	require.Equal(t, 1, repo.calls)
}
