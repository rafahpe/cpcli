// Copyright © 2017 Rafael Rivero
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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export some resource",
	Long: `Export a resource using the Web UI.

  - First argument is the resource name: "Service", "Devices", etc.
  - Second attribute is the password to protect the downloaded zip file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if fname, err := Singleton.Export(args); err != nil {
			Singleton.Log.Fatal(err)
		} else {
			fmt.Println("Resource", args[0], "exported to file", fname)
		}
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}
