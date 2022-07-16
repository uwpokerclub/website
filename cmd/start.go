package cmd

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the UWPSC Admin API server",
	Run: func(cmd *cobra.Command, args []string) {
		r := gin.Default()

		r.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		r.Run()
	},
}
