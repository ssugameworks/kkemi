package interfaces

import (
	"context"

	"github.com/ssugameworks/Discord-Bot/api"
)

// APIClient 외부 API와의 통신을 위한 인터페이스입니다
type APIClient interface {
	GetUserInfo(ctx context.Context, handle string) (*api.UserInfo, error)
	GetUserTop100(ctx context.Context, handle string) (*api.Top100Response, error)
	GetUserAdditionalInfo(ctx context.Context, handle string) (*api.UserAdditionalInfo, error)
	GetUserOrganizations(ctx context.Context, handle string) ([]api.Organization, error)
}
