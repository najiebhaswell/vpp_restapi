package main

import (
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "vpp-restapi/internal/api"
    "vpp-restapi/internal/handler/bond"
    _interface "vpp-restapi/internal/handler/_interface"
    "vpp-restapi/internal/handler/version"
    "vpp-restapi/internal/handler/lcpng"
    "vpp-restapi/internal/handler/vlan"
    "github.com/swaggo/gin-swagger"
    "github.com/swaggo/files"
    _ "vpp-restapi/docs"
    "vpp-restapi/internal/middleware"
)

func main() {
    log.Println("Starting VPP REST API server...")
    vppClient, err := api.NewVPPClient("/run/vpp/api.sock")
    if err != nil {
        log.Fatalf("Failed to connect to VPP: %v", err)
    }
    defer vppClient.Close()
    r := gin.Default()

    // PATCH: CORS CUSTOM - izinkan Authorization header
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: false,
        MaxAge: 12 * time.Hour,
    }))

    const myToken = "AexDQ4RyPi3jYETDHYFIxfFeQztzxBFoH3zZXGTTk0cI0RZqpzbqXM3epOeIOHik"
    auth := middleware.AuthMiddleware(myToken)

    // Swagger tanpa Auth
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API group dengan auth
    apiGroup := r.Group("/vpp", auth)
    version.RegisterRoutes(apiGroup, vppClient)
    _interface.RegisterRoutes(apiGroup, vppClient)
    _interface.RegisterConfigRoutes(apiGroup, vppClient)
    bond.RegisterRoutes(apiGroup, vppClient)
    lcpng.RegisterRoutes(apiGroup, vppClient)
    vlan.RegisterRoutes(apiGroup, vppClient)

    // DEBUG: Print all registered routes
    for _, ri := range r.Routes() {
        log.Printf("REGISTERED ROUTE: %s %s", ri.Method, ri.Path)
    }

    log.Println("Starting server on :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
