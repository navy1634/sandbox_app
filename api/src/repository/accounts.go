package repository

import (
	"context"

	"github.com/sandbox-nextjs/src/domain"
)

type AccountRepository interface {
	UpsertFromOIDCUser(ctx context.Context, user domain.OIDCUser) (domain.AppAccount, error)
}
