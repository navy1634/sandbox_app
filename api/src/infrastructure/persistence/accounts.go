package persistence

import (
	"context"
	"fmt"

	"github.com/sandbox-nextjs/src/domain"
	"github.com/sandbox-nextjs/src/ent"
)

type EntAccountRepository struct {
	client *ent.Client
}

func NewEntAccountRepository(client *ent.Client) *EntAccountRepository {
	return &EntAccountRepository{client: client}
}

func (r *EntAccountRepository) UpsertFromOIDCUser(ctx context.Context, user domain.OIDCUser) (domain.AppAccount, error) {
	if user.Subject == "" {
		return domain.AppAccount{}, fmt.Errorf("OIDC subject is required")
	}
	id, err := r.client.AppAccount.Create().
		SetOidcSubject(user.Subject).
		SetEmail(user.Email).
		OnConflictColumns("oidc_subject").
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
