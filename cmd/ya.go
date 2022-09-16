package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"youtube-audio/pkg/util/log"
)

var RootCmd = &cobra.Command{
	Use:     "ya",
	Short:   "Youtube Audio CLI",
	Long:    "CLI TOOL FOR SENDING YOUTUBE AUDIO TO TELEGRAM",
	Example: "ya run -m all",
}

type yaVersion struct {
	CliVersion string `json:"Cli version"`
}

var (
	yaVer     yaVersion
	logAsJSON bool
)

func Execute(version string) {
	RootCmd.Version = version

	yaVer = yaVersion{
		CliVersion: version,
	}

	cobra.OnInitialize(initConfig)

	setVersion()

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

}

func setVersion() {
	template := fmt.Sprintf(cliVersionTemplateString, yaVer.CliVersion)
	RootCmd.SetVersionTemplate(template)
}

func initConfig() {
	if logAsJSON {
		log.EnableJSONFormat()
	}

	viper.SetEnvPrefix("alicloud")
	viper.AutomaticEnv()
}

func init() {
	log.Infof("Init command ya ...")
	RootCmd.PersistentFlags().BoolVarP(&logAsJSON, "log-as-json", "", false, "Log output in JSON format")
}
