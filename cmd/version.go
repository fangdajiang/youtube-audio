package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"youtube-audio/pkg/util/log"
)

const cliVersionTemplateString = "CLI version: %s \n"

var output string

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Youtube Audio Cli version.",
	Example: `
# Version for Youtube Audio
ya version --output json
`,
	Run: func(cmd *cobra.Command, args []string) {
		if output != "" && output != "json" {
			_, err := fmt.Fprintf(os.Stdout, "An invalid output format was specified.")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			os.Exit(1)
		}
		switch output {
		case "":
			fmt.Printf(cliVersionTemplateString, yaVer.CliVersion)
		case "json":
			m, err := json.Marshal(yaVer)
			if err != nil {
				_, err2 := fmt.Fprintf(os.Stderr, err.Error())
				if err2 != nil {
					fmt.Printf("Error2: %v\n", err)
				}
				os.Exit(1)
			}
			fmt.Printf("%s\n", string(m))
		default:
			os.Exit(1)
		}
	},
}

func init() {
	log.Infof("Init command version ...")
	VersionCmd.Flags().BoolP("help", "h", false, "Print this help message")
	VersionCmd.Flags().StringVarP(&output, "output", "o", "", "The output format of the version command. Valid values are: json.")
	RootCmd.AddCommand(VersionCmd)
}
