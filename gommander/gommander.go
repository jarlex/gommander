package gommander

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var cnf *Config

//RootCmd root command of gommander
var RootCmd = &cobra.Command{
	Use:   "gommander",
	Short: "root comand of gommander",
	Long:  `This binary can make aceptation and stress test`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file wich contain a full plan of test")
}

func initConfig() {
	cnf = Read(cfgFile)
}
