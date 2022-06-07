package profile

import (
	"context"
	"coolcar/shared/id"
)

type Manager struct {

}

func (m *Manager) Verify(ctx context.Context,aid id.AccountId) (id.IdentityID,error)  {
	return id.IdentityID("identity1"), nil
}