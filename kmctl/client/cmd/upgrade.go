package cmd

import (
	"fmt"
	"sync"
	"time"

	"com.kubebackend/m/client/controller"
	"com.kubebackend/m/client/model"
	"github.com/spf13/cobra"
)

var upgradeType int
var upgradeVersion string
var upgradeYamlPath string

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Apply application upgrade yaml file to the all clusters",
	Long: `Apply application upgrade yaml file to the all clusters

For example:
upgrade -f <yaml-file-path>
upgrade -t <upgrade type> -v <version> -f <yaml-file-path>  # Version must be in the format of <00.00.00>
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Apply: %s\n", upgradeYamlPath)
		fmt.Println()

		var wg sync.WaitGroup
		for _, cluster := range clusters.Cluster {
			wg.Add(1)
			go func(cluster model.Cluster) {
				defer wg.Done()
				yamlCon := controller.NewYaml(&cluster.Host, &cluster.Port)
				err := yamlCon.UpgradeYaml(&upgradeType, &upgradeVersion, &upgradeYamlPath, &cluster)
				if err != nil {
					return
				}
				fmt.Println()
			}(cluster)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	today := time.Now()
	defaultVer := fmt.Sprintf("%d.%d.%d", today.Year()%100, today.Month(), today.Day())

	upgradeCmd.Flags().IntVarP(&upgradeType, "type", "t", 0, "Upgrade type 0: Micom Manager, 1: Device Bringup, 2: Navigation, 3: Middleware")
	upgradeCmd.Flags().StringVarP(&upgradeVersion, "version", "v", defaultVer, "Upgrade version")
	upgradeCmd.Flags().StringVarP(&upgradeYamlPath, "file", "f", "", "The yaml file path")

	upgradeCmd.MarkFlagRequired("type")
	upgradeCmd.MarkFlagRequired("file")
}
