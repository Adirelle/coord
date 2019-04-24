package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "coord",
	Short: "Coord is ...",
	Long:  "HTTP-based processus coordination.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Root().Help()
	},
}

var serverURL urlValue

func init() {
	viper.SetEnvPrefix("coord")
	viper.AutomaticEnv()

	serverURL.URL, _ = url.Parse("http://localhost:7500")
	rootCmd.PersistentFlags().VarP(&serverURL, "server", "s", "URL of coord server")
	viper.BindPFlag("server", rootCmd.Flags().Lookup("server"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type urlValue struct {
	*url.URL
}

func (v *urlValue) String() string {
	if v.URL != nil {
		return v.URL.String()
	}
	return ""
}

func (v *urlValue) Set(rawURL string) (err error) {
	v.URL, err = url.Parse(rawURL)
	return
}

func (v *urlValue) Type() string {
	return "URL"
}
