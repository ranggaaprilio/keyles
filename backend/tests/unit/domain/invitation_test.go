package domain_test

import (
	"testing"
	"time"

	"github.com/ranggaaprilio/keyles/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestInvitationIsExpired_PastExpiry(t *testing.T) {
	inv := &entities.Invitation{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, inv.IsExpired())
}

func TestInvitationIsExpired_FutureExpiry(t *testing.T) {
	inv := &entities.Invitation{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, inv.IsExpired())
}

func TestInvitationIsAccepted_AcceptedStatus(t *testing.T) {
	inv := &entities.Invitation{
		Status: entities.InvitationStatusAccepted,
	}
	assert.True(t, inv.IsAccepted())
}

func TestInvitationIsAccepted_PendingStatus(t *testing.T) {
	inv := &entities.Invitation{
		Status: entities.InvitationStatusPending,
	}
	assert.False(t, inv.IsAccepted())
}

func TestInvitationIsAccepted_ExpiredStatus(t *testing.T) {
	inv := &entities.Invitation{
		Status: entities.InvitationStatusExpired,
	}
	assert.False(t, inv.IsAccepted())
}

func TestInvitationStatusConstants(t *testing.T) {
	assert.Equal(t, entities.InvitationStatus("pending"), entities.InvitationStatusPending)
	assert.Equal(t, entities.InvitationStatus("accepted"), entities.InvitationStatusAccepted)
	assert.Equal(t, entities.InvitationStatus("expired"), entities.InvitationStatusExpired)
}

func TestInvitationTTL(t *testing.T) {
	assert.Equal(t, 72*time.Hour, entities.InvitationTTL)
}
