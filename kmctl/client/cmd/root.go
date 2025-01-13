package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"com.kubebackend/m/client/model"
)

var (
	cfgFile  string
	clusters model.Clusters

	rootCmd = &cobra.Command{
		Use:   "client",
		Short: "This is a client for the kube multi cli",
		Long: `This is a client for the kube multi cli.

You need to server to use this client.
Server is a gRPC server that provides information about Kubernetes resources.

You can get information about nodes and pods from the server.
You can also apply and delete kubernetes yaml file to the clusters.
		
You make config.yaml file in $HOME/.config/kmctl directory.
You can set multiple clusters in the config file.

Example config.yaml:
		
server:
- name: localhost
  port: 50051
  host: 127.0.0.1
- name: cluster1
  port: 50051
  host: 192.168.5.10
- name: cluster2
  port: 50051
  host: 192.168.5.11
 ...
`,
		Run: func(cmd *cobra.Command, args []string) {
			version, _ := cmd.Flags().GetBool("version")
			if version {
				fmt.Println("Kube Multi CLI")
				fmt.Printf("Version: %s\n", VERSION)
			} else {
				cmd.Help()
			}
		},
	}
)

const VERSION = "0.1.4"

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/kmctl/config.yaml)")
	rootCmd.Flags().BoolP("version", "v", false, "Print the version of the client")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(deleteCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("$HOME/.config/kmctl")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(&clusters); err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}
	} else {
		log.Fatalf("Failed to read config: %v", err)
	}
}
