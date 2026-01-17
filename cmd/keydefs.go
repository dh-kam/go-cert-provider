package cmd

type contextKey interface{}

var (
	KeyForOptions      contextKey = "options"
	KeyForGin          contextKey = "gin"
	KeyForJwtSecret    contextKey = "jwt_secret_key" //nolint:gosec // not a credential
	KeyForCertRegistry contextKey = "cert_registry"
)
