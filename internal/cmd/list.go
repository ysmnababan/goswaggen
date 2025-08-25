package cmd

import (
	"fmt"
	"os"

	"github.com/ysmnababan/goswaggen/internal/parser"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "controller lists",
	Long:    "list all the controller grouped by its package",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("All Controllers")
		var p IParser
		var err error
		p, err = parser.NewParser(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while fetch all the handlers: %v\n", err)
			os.Exit(1)
		}
		c := p.GetAllHandlers()
		for p, funcs := range c {
			fmt.Println(p, ":")
			for _, f := range *funcs {
				fmt.Println("	", f)
			}
			fmt.Printf("\n")
		}
	},
}

func GetAllControllers() map[string][]string {
	return map[string][]string{
		"user":  {"GetUser", "CreateUser", "UpdateUser", "DeleteUser"},
		"order": {"CreateOrder", "GetCart", "CancelOrder"},
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
