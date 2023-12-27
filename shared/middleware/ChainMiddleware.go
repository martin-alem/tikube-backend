package shared_middleware

import "tikube-backend/shared/utils"

// ChainMiddlewares applies all provided middlewares to a given handler
func ChainMiddlewares(handler utils.HTTPHandler, middlewares ...utils.Middleware) utils.HTTPHandler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
