// Copyright © 2024 JR team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"
	"github.com/jrnd-io/jrv2/pkg/constants"
	"github.com/spf13/cobra"
)

var Version = "DEV"
var GoVersion string
var BuildTime string
var BuildUser string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints JR version number",
	Long:  `prints JR version number`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("JR System Dir: %s\n", constants.JrSystemDir)
		fmt.Printf("JR User Dir  : %s\n", constants.JrUserDir)
		fmt.Printf("JR Version   : %s\n", Version)
		fmt.Printf("Built with   : %s\n", GoVersion)
		fmt.Printf("By           : %s\n", BuildUser)
		fmt.Printf("At           : %s\n", BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
