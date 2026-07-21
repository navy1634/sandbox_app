package persistence

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sandbox-nextjs/src/domain"
	"github.com/sandbox-nextjs/src/ent"
)

type EntAccountRepository struct {
	client *ent.Client
}

func NewEntAccountRepository(client *ent.Client) *EntAccountRepository {
	return &EntAccountRepository{client: client}
}

func (r *EntAccountRepository) UpsertFromAuthUser(ctx context.Context, user domain.AuthUser) (domain.AppAccount, error) {
	if user.AccountID > math.MaxInt || user.AccountID < 1 {
		return domain.AppAccount{}, fmt.Errorf("auth account id is out of range: %d", user.AccountID)
	}
	accountID := int(user.AccountID)
	now := time.Now()
	id, err := r.client.AppAccount.Create().
		SetID(accountID).
		SetEmail(user.Email).
		SetName(user.Name).
		SetPicture(user.Picture).
		SetUpdatedAt(now).
		OnConflictColumns("id").
		UpdateNewValues().
		ID(ctx)
	if err != nil {
		return domain.AppAccount{}, err
	}

	account, err := r.client.AppAccount.Get(ctx, id)
	if err != nil {
		return domain.AppAccount{}, err
	}
	return toDomainAppAccount(account), nil
}

func toDomainAppAccount(account *ent.AppAccount) domain.AppAccount {
	return domain.AppAccount{
		ID:        account.ID,
		Email:     account.Email,
		Name:      account.Name,
		Picture:   account.Picture,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}
