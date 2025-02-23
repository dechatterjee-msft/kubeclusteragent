package linux

import (
	"context"
)

type PackageManagerFactory interface {
	CheckInstalled(ctx context.Context, packageName string) bool
	Install(ctx context.Context, packageNames ...string) error
	Update(ctx context.Context) error
	AddKey(ctx context.Context, urlStr string) error
	AddRepository(ctx context.Context, repository, filename string) error
	Uninstall(ctx context.Context, packageNames ...string) error
	RemoveRepository(ctx context.Context, repository, filename string) error
}
