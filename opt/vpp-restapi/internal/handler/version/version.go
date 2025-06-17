package version

import (
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "vpp-restapi/binapi/vpe"
    "vpp-restapi/internal/api"
)

// RegisterRoutes sets up the version-related HTTP routes.
func RegisterRoutes(r gin.IRoutes, vppClient *api.VPPClient) {
    r.GET("/version", getVersionHandler(vppClient))
}

// @Summary Show VPP version
// @Description Get the running VPP version and build date
// @Tags version
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /vpp/version [get]
func getVersionHandler(vppClient *api.VPPClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        log.Println("Handling /vpp/version request")
        ch, err := vppClient.NewAPIChannel()
        if err != nil {
            log.Printf("Failed to create API channel: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "channel creation failed", "details": err.Error()})
            return
        }
        defer ch.Close()

        req := &vpe.ShowVersion{}
        reply := &vpe.ShowVersionReply{}
        if err := ch.SendRequest(req).ReceiveReply(reply); err != nil {
            log.Printf("API request failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed", "details": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "version":    reply.Version,
            "build_date": reply.BuildDate,
        })
    }
}
