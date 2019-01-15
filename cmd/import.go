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
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import some resource",
	Long: `Import a resource using the Web UI.

  - First argument is the path of the file to import
  - Second argument is the resource name: "Service", "Devices", etc.
  - Third argument is the password to protect the downloaded zip file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := Singleton.Import(args); err != nil {
			Singleton.Log.Fatal(err)
		} else {
			fmt.Println("Resource", args[1], "from file", args[0])
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)
}
