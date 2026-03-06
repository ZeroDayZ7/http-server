package di

import (
	"github.com/zerodayz7/http-server/internal/handler"
	"github.com/zerodayz7/http-server/internal/worker"
)

type InteractionModule struct {
	Handler *handler.InteractionHandler
	Worker  *worker.InteractionWorker
}

func NewInteractionModule(h *handler.InteractionHandler, w *worker.InteractionWorker) *InteractionModule {
	return &InteractionModule{
		Handler: h,
		Worker:  w,
	}
}
