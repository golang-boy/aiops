/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"copilot/k8s"

	"github.com/spf13/cobra"
)

// eventCmd represents the event command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		k8s, err := k8s.NewK8SHelper(kubeconfig)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		eventLog, err := k8s.GetPodEventsAndLogs()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		result, err := k8s.AskGpt(eventLog)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(result)
	},
}

func init() {
	pprofCmd.AddCommand(eventCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
