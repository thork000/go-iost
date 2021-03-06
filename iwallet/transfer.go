// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iwallet

import (
	"github.com/spf13/cobra"
)

var memo string

var transferCmd = &cobra.Command{
	Use:     "transfer receiver amount",
	Aliases: []string{"trans"},
	Short:   "Transfer IOST",
	Long:    `Transfer IOST`,
	Example: `  iwallet transfer test1 100 --account test0
  iwallet transfer test1 100 --account test0 --memo "just for test :D\n中文测试\n😏"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "receiver", "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[1], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return saveOrSendAction("token.iost", "transfer", "iost", accountName, args[0], args[1], memo)
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)
	transferCmd.Flags().StringVarP(&memo, "memo", "", "", "memo of transfer")
}
