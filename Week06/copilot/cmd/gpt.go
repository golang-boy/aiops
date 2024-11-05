/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"copilot/k8s"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// gptCmd represents the gpt command
var gptCmd = &cobra.Command{
	Use:   "gpt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		startChat()
	},
}

func init() {
	askCmd.AddCommand(gptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func startChat() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("有什么可以帮助你：")

	k8s, err := k8s.NewK8SHelper(kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	k8s.SetTools()

	for {
		fmt.Print(">>>> ")
		if scanner.Scan() {
			input := scanner.Text()
			if input == "exit" {
				fmt.Println("┏(＾0＾)┛!")
				break
			}
			if input == "" {
				continue
			}
			response := k8s.FuncCalling(input)
			fmt.Println(response)
		}
	}
}
