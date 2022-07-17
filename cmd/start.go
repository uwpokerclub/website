package cmd

import (
	"api/internal/database"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type Login struct {
	Username string
	Password string
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the UWPSC Admin API server",
	Run: func(cmd *cobra.Command, args []string) {
		r := gin.Default()

		db, err := database.OpenConnection()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open connection to the database: %s", err.Error())
			os.Exit(1)
		}

		res := db.Create(Login{Username: "adam", Password: "password"})
		if err := res.Error; err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create the login: %s", err.Error())
			os.Exit(1)
		}

		r.GET("/ping", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})

		r.Run()
	},
}
