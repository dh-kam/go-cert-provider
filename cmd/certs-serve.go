package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dh-kam/go-cert-provider/config"
	"github.com/dh-kam/go-cert-provider/graph"
	"github.com/dh-kam/go-cert-provider/graph/generated"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GraphQL server",
	Long: `Start the HTTP server with GraphQL API for certificate retrieval.

The server provides:
- GraphQL API endpoint at /graphql
- GraphQL Playground at /
- Health check endpoint at /health

Examples:
  # Start server with default settings
  go-cert-provider certs serve

  # Start server on custom port
  go-cert-provider certs serve --listen-port 8080

  # Start with Porkbun provider
  go-cert-provider certs serve \
    --porkbun-api-key "your-key" \
    --porkbun-secret-key "your-secret" \
    --porkbun-domains "example.com,test.com"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		listenPort, err := cmd.Flags().GetInt("listen-port")
		if err != nil {
			return err
		}
		listenAddr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			return err
		}
		jwtSecretKey, err := cmd.Flags().GetString("jwt-secret-key")
		if err != nil {
			return err
		}

		if appState == nil {
			return fmt.Errorf("certificate system not initialized")
		}

		providerRegistry := appState.providerRegistry
		bootstrapManager := appState.bootstrapManager

		if jwtSecretKey == "" {
			jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
		}
		if jwtSecretKey == "" {
			return fmt.Errorf(`JWT secret key is required for server operation.

The server uses JWT tokens for authentication. Without a secret key, 
the server cannot verify JWT tokens and would be non-functional.

Please provide a JWT secret key using one of these methods:

  1. Environment variable:
     export JWT_SECRET_KEY="your-secret-key"
     
  2. Command line flag:
     --jwt-secret-key "your-secret-key"
     
  3. Generate a new secret key:
     go-cert-provider jwt create-secret-key

Then start the server with the generated key.`)
		}

		// Validate that we have at least one domain to manage
		domains := providerRegistry.ListDomains()
		if len(domains) == 0 {
			return fmt.Errorf(`no domains available for certificate management.

The server requires at least one domain to be configured. 
Please configure a provider with domains using one of these methods:

  1. Environment variables (Porkbun example):
     export PORKBUN_API_KEY="your-api-key"
     export PORKBUN_SECRET_KEY="your-secret-key"
     # Optional: specify domains manually
     export PORKBUN_DOMAINS="example.com,*.example.com"
     
  2. Command line flags:
     --porkbun-api-key "your-api-key" \
     --porkbun-secret-key "your-secret-key" \
     --porkbun-domains "example.com,test.com"
     
  3. Auto-discovery (Porkbun):
     If you provide only API credentials without specifying domains,
     the system will automatically discover all active domains from
     your Porkbun account.

For more information, see: go-cert-provider domain list --help`)
		}

		fmt.Printf("Configured providers: %v\n", bootstrapManager.GetConfiguredProviders())
		fmt.Printf("Managed domains: %v\n", domains)
		fmt.Printf("JWT authentication: enabled\n")

		serverConfig := config.NewServerConfig()
		if listenPort != 0 {
			serverConfig.SetPort(listenPort)
		}
		if listenAddr != "" {
			serverConfig.SetAddr(listenAddr)
		}

		router := gin.Default()

		// GraphQL playground
		router.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/graphql")))

		// GraphQL endpoint
		gqlHandler := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
		gqlHandler.AddTransport(transport.POST{})
		gqlHandler.Use(extension.Introspection{})

		// Custom middleware to add gin context, JWT secret, and provider registry to GraphQL context
		router.POST("/graphql", func(c *gin.Context) {
			// Add gin context, JWT secret key, and provider registry to the request context
			ctx := context.WithValue(c.Request.Context(), KeyForGin, c)
			ctx = context.WithValue(ctx, KeyForJwtSecret, jwtSecretKey)
			ctx = context.WithValue(ctx, KeyForCertRegistry, providerRegistry)
			c.Request = c.Request.WithContext(ctx)

			// Call the GraphQL handler
			gin.WrapH(gqlHandler)(c)
		})

		// Health check endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"version":   config.Version,
				"providers": bootstrapManager.GetConfiguredProviders(),
				"domains":   providerRegistry.ListDomains(),
			})
		})

		srv := &http.Server{
			Addr:              serverConfig.GetListenAddr(),
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan

			fmt.Println("\nShutting down server...")
			if err := srv.Shutdown(context.Background()); err != nil {
				fmt.Printf("Server forced to shutdown: %v\n", err)
			}
			fmt.Println("Server exiting")
		}()


		fmt.Printf("Server starting on %s\n", serverConfig.GetListenAddr())
		fmt.Printf("GraphQL Playground: http://%s/\n", serverConfig.GetListenAddr())
		fmt.Printf("GraphQL Endpoint: http://%s/graphql\n", serverConfig.GetListenAddr())
		fmt.Printf("Health Check: http://%s/health\n", serverConfig.GetListenAddr())

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("failed to start server: %v", err)
		}

		return nil
	},
}

func init() {
	flags := serveCmd.Flags()
	flags.Int("listen-port", 0, "Port to listen on (overrides LISTEN_PORT env var)")
	flags.String("listen-addr", "", "Address to listen on (overrides LISTEN_ADDR env var)")
	flags.String("jwt-secret-key", "", "JWT secret key for token verification (overrides JWT_SECRET_KEY env var)")

	certsCmd.AddCommand(serveCmd)
}
