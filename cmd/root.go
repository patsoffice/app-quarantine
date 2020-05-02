// Copyright © 2020 Patrick Lawrence <patrick.lawrence@gmail.com>
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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/xattr"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "app-quarantine",
		Short: "Identify (and optionally fix) quarantined applications on macOS",
		Run:   rootRun,
	}
	rootCmdFlags = struct {
		appPaths        []string
		appRegex        string
		fix             bool
		quarantineXattr string
	}{}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().StringSliceVarP(&rootCmdFlags.appPaths, "application-path", "p", []string{"/Applications"}, "Application path(s)")
	rootCmd.Flags().BoolVarP(&rootCmdFlags.fix, "fix", "f", false, "Remove quarantine attribute from application?")
	rootCmd.Flags().StringVar(&rootCmdFlags.appRegex, "app-regex", `\.app$`, "A regular expression to match applications")
	rootCmd.Flags().StringVar(&rootCmdFlags.quarantineXattr, "quarantine-xattr", "com.apple.quarantine", "The xattr to match")
}

func stringInSlice(a string, b []string) bool {
	for _, v := range b {
		if a == v {
			return true
		}
	}
	return false
}

func hasQuarantineAttr(path, qxattr string) (bool, error) {
	attrs, err := xattr.LList(path)
	if err != nil {
		return false, err
	}

	if stringInSlice(qxattr, attrs) {
		return true, nil
	}
	return false, nil
}

func findApps(appPath string, appRE regexp.Regexp, qxattr string) ([]string, error) {
	apps := []string{}

	files, err := ioutil.ReadDir(appPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		fn := file.Name()
		if appRE.MatchString(fn) {
			path := filepath.Join(appPath, fn)
			ok, err := hasQuarantineAttr(path, qxattr)
			if err != nil {
				return nil, err
			}
			if ok {
				apps = append(apps, path)
			}
		}
	}

	return apps, nil
}

func removeQuarantineXattr(app, qxattr string) error {
	return xattr.LRemove(app, qxattr)
}

func rootRun(cobraCmd *cobra.Command, args []string) {
	appRE, err := regexp.Compile(rootCmdFlags.appRegex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling regular expression: %v\n", err)
		os.Exit(-1)
	}

	toFix := []string{}

	for _, appPath := range rootCmdFlags.appPaths {
		apps, err := findApps(appPath, *appRE, rootCmdFlags.quarantineXattr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding applications: %v\n", err)
			os.Exit(-1)
		}
		toFix = append(toFix, apps...)
	}

	if len(toFix) == 0 {
		fmt.Println("No applications are quarantined.")
	} else {
		for _, app := range toFix {
			fmt.Printf("%s is quarantined", app)
			if rootCmdFlags.fix {
				if err := removeQuarantineXattr(app, rootCmdFlags.quarantineXattr); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				fmt.Printf("... Removed quarantine.")
			}
			fmt.Println()
		}
	}
}
