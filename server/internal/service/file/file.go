package file

import (
	"github.com/zhian9/GoForge/server/internal/service/file/repository"
)

// ServiceContext 服务上下文
type ServiceContext struct {
	Config   Config
	FileRepo repository.FileRepository
}

// NewServiceContext 创建服务上下文
func NewServiceContext(c Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	ctx.FileRepo = repository.NewFileRepository(c.Storage)

	return ctx
}
