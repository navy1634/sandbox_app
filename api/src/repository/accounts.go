package repository

import (
	"context"

	"github.com/sandbox-nextjs/src/domain"
)

type AccountRepository interface {
	UpsertFromAuthUser(ctx context.Context, user domain.AuthUser) (domain.AppAccount, error)
}
