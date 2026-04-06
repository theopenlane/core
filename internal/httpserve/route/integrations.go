package route

func integrationsEnabled(router *Router) bool {
	return router != nil && router.Handler != nil && router.Handler.IntegrationsRuntime != nil
}
