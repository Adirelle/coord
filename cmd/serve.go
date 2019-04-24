package cmd

import (
	"log"
	"os"

	lib "github.com/Adirelle/coord/lib"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a standalone server",
	Long:  `Run a standalone server.`,
	Run:   runServe,
	Args:  cobra.NoArgs,
}

func runServe(cmd *cobra.Command, args []string) {
	l := log.New(os.Stdout, "State: ", log.LstdFlags)
	state := lib.NewLocalState(l)

	router := gin.Default()

	lib.MakeServer(state, router.Group(serverURL.Path))

	err := router.Run(serverURL.Host)
	if err != nil {
		log.Fatal(err.Error())
	}
}
