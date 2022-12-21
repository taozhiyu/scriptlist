package script

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/codfrm/cago/database/redis"
	"github.com/codfrm/cago/pkg/limit"
	"github.com/codfrm/cago/pkg/utils/httputils"
	"github.com/gin-gonic/gin"
	api "github.com/scriptscat/scriptlist/internal/api/script"
	"github.com/scriptscat/scriptlist/internal/model"
	service "github.com/scriptscat/scriptlist/internal/service/script"
	"github.com/scriptscat/scriptlist/internal/service/user"
)

type Script struct {
	limit *limit.PeriodLimit
}

func NewScript() *Script {
	return &Script{
		limit: limit.NewPeriodLimit(
			300, 10, redis.Default(), "limit:create:script",
		),
	}
}

// List 获取脚本列表
func (s *Script) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	return service.Script().List(ctx, req)
}

// Create 创建脚本/库
func (s *Script) Create(ctx context.Context, req *api.CreateRequest) (*api.CreateResponse, error) {
	cancel, err := s.limit.Take(ctx, strconv.FormatInt(user.Auth().Get(ctx).UID, 10))
	if err != nil {
		return nil, err
	}
	resp, err := service.Script().Create(ctx, req)
	if err != nil {
		if err := cancel(); err != nil {
			return nil, err
		}
		return nil, err
	}
	return resp, nil
}

// UpdateCode 更新脚本/库代码
func (s *Script) UpdateCode(ctx context.Context, req *api.UpdateCodeRequest) (*api.UpdateCodeResponse, error) {
	cancel, err := s.limit.Take(ctx, strconv.FormatInt(user.Auth().Get(ctx).UID, 10))
	if err != nil {
		return nil, err
	}
	resp, err := service.Script().UpdateCode(ctx, req)
	if err != nil {
		if err := cancel(); err != nil {
			return nil, err
		}
		return nil, err
	}
	return resp, nil
}

// MigrateEs 全量迁移数据到es
func (s *Script) MigrateEs(ctx context.Context, req *api.MigrateEsRequest) (*api.MigrateEsResponse, error) {
	if user.Auth().Get(ctx).AdminLevel != model.Admin {
		return nil, httputils.NewError(http.StatusForbidden, -1, "无权限")
	}
	go service.Script().MigrateEs()
	return &api.MigrateEsResponse{}, nil
}

func (s *Script) Download() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.HasSuffix(ctx.Request.URL.Path, ".user.js") || strings.HasSuffix(ctx.Request.URL.Path, ".user.sub.js") {
			s.downloadScript(ctx)
		} else if strings.HasSuffix(ctx.Request.URL.Path, ".meta.js") {
			s.getScriptMeta(ctx)
		} else {
			ctx.AbortWithStatus(http.StatusNotFound)
		}
	}
}
