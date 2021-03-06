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
	"strconv"

	"github.com/spf13/cobra"
)

var other string

var buyCmd = &cobra.Command{
	Use:     "ram-buy amount",
	Aliases: []string{"buy"},
	Short:   "Buy ram from system",
	Long:    `Buy ram from system`,
	Example: `  iwallet sys buy 100 --account test0
  iwallet sys buy 100 --account test0 --ram_receiver test1`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "amount"); err != nil {
			return err
		}
		if err := checkFloat(cmd, args[0], "amount"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if other == "" {
			other = accountName
		}
		amount, _ := strconv.ParseFloat(args[0], 64)
		return saveOrSendAction("ram.iost", "buy", accountName, other, amount)
	},
}

var sellCmd = &cobra.Command{
	Use:     "ram-sell amount",
	Aliases: []string{"sell"},
	Short:   "Sell unused ram to system",
	Long:    `Sell unused ram to system`,
	Example: `  iwallet sys sell 100 --account test0
  iwallet sys sell 100 --account test0 --token_receiver test1`,
	Args: buyCmd.Args,
	RunE: func(cmd *cobra.Command, args []string) error {
		if other == "" {
			other = accountName
		}
		amount, _ := strconv.ParseFloat(args[0], 64)
		return saveOrSendAction("ram.iost", "sell", accountName, other, amount)
	},
}

var rtransCmd = &cobra.Command{
	Use:     "ram-transfer receiver amount",
	Aliases: []string{"ram-trans", "rtrans"},
	Short:   "Transfer ram",
	Long:    `Transfer ram`,
	Example: `  iwallet sys ram-transfer test1 100 --account test0`,
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
		amount, _ := strconv.ParseFloat(args[1], 64)
		return saveOrSendAction("ram.iost", "lend", accountName, args[0], amount)
	},
}

func init() {
	systemCmd.AddCommand(buyCmd)
	buyCmd.Flags().StringVarP(&other, "ram_receiver", "", "", "who gets the bought ram")
	systemCmd.AddCommand(sellCmd)
	sellCmd.Flags().StringVarP(&other, "token_receiver", "", "", "who gets the returned IOST after selling")
	systemCmd.AddCommand(rtransCmd)
}
