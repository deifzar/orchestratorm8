/*
Copyright © 2024 i@deifzar.me

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"deifzar/orchestratorm8/pkg/log8"
	om8 "deifzar/orchestratorm8/pkg/orchestratorm8"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "orchestratorm8",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create log and tmp directories if they don't exist
		if err := os.MkdirAll("log", 0755); err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Failed to create log directory")
			return err
		}
		if err := os.MkdirAll("tmp", 0755); err != nil {
			log8.BaseLogger.Error().Err(err).Msg("Failed to create tmp directory")
			return err
		}

		log8.BaseLogger.Info().Msg("Initilising OrchestratorM8 ...")
		o, err := om8.NewOrchestratorM8()
		if err != nil {
			log8.BaseLogger.Fatal().Msg("Error in Cobra CLI execute. Something wrong with the DB or RabbitMQ server connections.")
			return err
		}
		err = o.InitOrchestrator()
		if err != nil {
			log8.BaseLogger.Fatal().Msg("Error in Cobra CLI execute. Something wrong with declaring the RabbitMQ Exchanges.")
			return err

		}
		log8.BaseLogger.Info().Msg("OrchestratorM8 declared successfully all the RabbitMQ Exchanges and the ASMM8 queue.")
		log8.BaseLogger.Info().Msg("OrchestratorM8 started ...")
		o.StartOrchestrator()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log8.BaseLogger.Fatal().Msg("Error in Cobra CLI execute. Exit(1)")
		log8.BaseLogger.Fatal().Msg(err.Error())
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.orchestratorm8.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
