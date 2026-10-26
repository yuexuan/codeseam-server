package service

import "context"

// Dedicated service revocation must not wait for the subscription L1 TTL.
func (s *SubscriptionService) GetActiveSubscriptionFresh(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	return s.userSubRepo.GetByUserIDAndGroupID(ctx, userID, groupID)
}
