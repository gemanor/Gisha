/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmds

import (
	"fmt"

	gisha "github.com/gemanor/gisha/pkg/login"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the service",
	Run: func(cmd *cobra.Command, args []string) {
		options := gisha.LoginOptions{
			AppID:        cmd.Flag("app-id").Value.String(),
			ClientID:     cmd.Flag("client-id").Value.String(),
			ClientSecret: cmd.Flag("client-secret").Value.String(),
			Scopes:       cmd.Flag("scopes").Value.String(),
			AuthURL:      cmd.Flag("auth-url").Value.String(),
			TokenURL:     cmd.Flag("token-url").Value.String(),
		}
		token, err := gisha.Login(options)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(token.AccessToken)
	},
}

func init() {
	loginCmd.Flags().StringP("app-id", "a", "", "Application ID")
	loginCmd.Flags().StringP("client-id", "c", "", "Client ID")
	loginCmd.MarkFlagRequired("client-id")
	loginCmd.Flags().StringP("client-secret", "s", "", "Client secret")
	loginCmd.MarkFlagRequired("client-secret")
	loginCmd.Flags().StringP("scopes", "S", "", "Authorization scopes")
	loginCmd.Flags().StringP("auth-url", "u", "", "Authorization URL")
	loginCmd.MarkFlagRequired("auth-url")
	loginCmd.Flags().StringP("token-url", "t", "", "Token URL")
	loginCmd.MarkFlagRequired("token-url")
	rootCmd.AddCommand(loginCmd)
}
