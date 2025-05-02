package commands

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/http"
)

func Execute() {
	rootCmd := &cobra.Command{
		Use:   "smokey",
		Short: "manage proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "manage server",
	}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "start server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}
			return http.Start(cfg, getLogFormatter(args))
		},
	}
	serverCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serverCmd)
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func getLogFormatter(args []string) logrus.Formatter {
	if len(args) > 0 && args[0] == "text-formatter" {
		return &logrus.TextFormatter{}
	}
	return &logrus.JSONFormatter{}
}
