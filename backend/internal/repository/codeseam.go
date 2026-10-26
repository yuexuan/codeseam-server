package repository

import (
	"context"
	dbaccount "github.com/Wei-Shaw/sub2api/ent/account"
)

// Check exclusivity without loading other customers' credentials into memory.
func (r *accountRepository) CodeSeamProxyExclusive(ctx context.Context, proxyID, accountID int64) (bool, error) {
	exists, err := r.client.Account.Query().Where(dbaccount.ProxyIDEQ(proxyID), dbaccount.IDNEQ(accountID)).Exist(ctx)
	return !exists, err
}
