package main

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

const (
	exampleInfo = "\tkeentune profile info --name cpu_high_load.conf"
	exampleSet  = "\tkeentune profile set --group1 cpu_high_load.conf\n" +
		"\tkeentune profile set cpu_high_load.conf"
	exampleGenerate     = "\tkeentune profile generate --name tune_test.conf --output gen_param_test.json"
	exampleProfDelete   = "\tkeentune profile delete --name tune_test.conf"
	exampleProfList     = "\tkeentune profile list"
	exampleProfRollback = "\tkeentune profile rollback"
)

// createProfileCmds ...
func createProfileCmds() *cobra.Command {
	var profCmd = &cobra.Command{
		Use:   "profile [command]",
		Short: "Static tuning with expert profiles",
		Long:  "Static tuning with expert profiles",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
			exampleProfDelete,
			exampleGenerate,
			exampleInfo,
			exampleProfList,
			exampleProfRollback,
			exampleSet,
		),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			cmd.Usage()
			os.Exit(1)
		},
	}

	var profileCommands []*cobra.Command
	profileCommands = append(profileCommands, decorateCmd(infoCmd()))
	profileCommands = append(profileCommands, decorateCmd(setCmd()))
	profileCommands = append(profileCommands, decorateCmd(deleteProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(listProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(rollbackCmd("profile")))
	profileCommands = append(profileCommands, decorateCmd(generateCmd()))

	profCmd.AddCommand(profileCommands...)
	return profCmd
}

// keentune profile info
func infoCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Show information of the specified profile",
		Long:    "Show information of the specified profile",
		Example: exampleInfo,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			} else {
				name = strings.TrimSuffix(name, ".conf") + ".conf"
				RunInfoRemote(name)
			}
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name, query by command \"keentune profile list\"")
	return cmd
}

// keentune profile list
func listProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all profiles",
		Long:    "List all profiles",
		Example: exampleProfList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote("profile")
		},
	}
	return cmd
}

func setCmd() *cobra.Command {
	var setFlag SetFlag
	const GroupNum int = 20
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Apply a profile to the target machine",
		Long:    "Apply a profile to the target machine",
		Example: exampleSet,
		Run: func(cmd *cobra.Command, args []string) {
			if (len(args) == 0 && setWithoutAnyGroup(setFlag.ConfFile)) ||
				(len(args) > 0 && len(strings.Trim(args[0], " ")) == 0) {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			}

			// bind configuration file to group
			bindFileToGroup(args, setFlag)

			RunSetRemote(setFlag)
			return
		},
	}

	var group string
	if err := config.InitTargetGroup(); err != nil {
		setFlag.Group = make([]bool, GroupNum)
		setFlag.ConfFile = make([]string, GroupNum)
		for index := 0; index < GroupNum; index++ {
			group = fmt.Sprintf("group%d", index)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}
	} else {
		setFlag.Group = make([]bool, len(config.KeenTune.Target.Group))
		setFlag.ConfFile = make([]string, len(config.KeenTune.Target.Group))
		for index, _ := range config.KeenTune.Target.Group {
			group = fmt.Sprintf("group%d", index+1)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}
	}

	return cmd
}

func setWithoutAnyGroup(groupFiles []string) bool {
	for _, fileName := range groupFiles {
		if len(strings.Trim(fileName, " ")) != 0 {
			return false
		}
	}

	return true
}

func bindFileToGroup(args []string, setFlag SetFlag) {
	// Case1: bind all groups to the same configuration, when args passed. 
	if len(args) > 0 {
		if !strings.HasSuffix(strings.Trim(args[0], " "), ".conf") {
			fmt.Printf("%v Invalid value: '%v' is not with '.conf' suffix.\n", ColorString("red", "[ERROR]"), args[0])
			os.Exit(1)
		}

		filePath, err := checkProfile(args[0])
		if err != nil {
			fmt.Printf("%v %v\n", ColorString("red", "[ERROR]"), err)
			os.Exit(1)
		}

		for i, _ := range setFlag.ConfFile {
			setFlag.Group[i] = true
			setFlag.ConfFile[i] = filePath
		}

		return
	}

	// Case2: bind a group according to the corresponding configuration by '--groupx' flag.
	for i, v := range setFlag.ConfFile {
		if len(v) != 0 {
			if !strings.HasSuffix(v, ".conf") {
				fmt.Printf("%v Invalid value: group%v, '%v' is not with '.conf' suffix.\n", ColorString("red", "[ERROR]"), i+1, v)
				os.Exit(1)
			}

			filePath, err := checkProfile(v)
			if err != nil {
				fmt.Printf("%v %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			setFlag.Group[i] = true
			setFlag.ConfFile[i] = filePath
		}
	}

	return
}

func checkProfile(fileName string) (string, error) {
	var exactMatch = true
	fullPath, homeFileList, _ := file.WalkFilePath(config.GetProfileHomePath(""), fileName, exactMatch)
	dir, _ := path.Split(fileName)
	// A file with the same name appears and no parent directory is specified
	if len(homeFileList) > 1 && len(strings.TrimSpace(dir)) == 0 {
		return "", fmt.Errorf("file '%v' with the same name exists in %v.\n \tPlease specify a more detailed path to execute again", fileName, strings.Join(fullPath, ", "))
	}

	filePath := config.GetProfilePath(fileName)
	if filePath == "" {
		return "", fmt.Errorf("file '%v' does not exist, expect in '%v' nor in '%v'", fileName,
			fmt.Sprintf("%s/profile", config.KeenTune.Home),
			fmt.Sprintf("%s/profile", config.KeenTune.DumpHome))
	}

	return filePath, nil
}

func deleteProfileCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a profile",
		Long:    "Delete a profile",
		Example: exampleProfDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			}

			flag.Cmd = "profile"
			flag.Name = strings.TrimSuffix(flag.Name, ".conf") + ".conf"

			initWorkDirectory()
			WorkPath := config.GetProfileWorkPath(flag.Name)
			_, err := os.Stat(WorkPath)
			if err == nil {
				fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
				if !confirm() {
					fmt.Println("[-] Give Up Delete")
					return
				}
				RunDeleteRemote(flag)
			} else {
				HomePath := config.GetProfileHomePath(flag.Name)
				_, err = os.Stat(HomePath)
				if err == nil {
					fmt.Printf("%v File '%v' is not supported to delete\n", ColorString("red", "[ERROR]"), HomePath)
				} else {
					fmt.Printf("%v File '%v' does not exist.\n", ColorString("red", "[ERROR]"), flag.Name)
				}
				os.Exit(1)
			}

			return
		},
	}

	cmd.Flags().StringVar(&flag.Name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func generateCmd() *cobra.Command {
	var genFlag GenFlag
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a parameter configuration file from profile",
		Long:    "Generate a parameter configuration file from profile",
		Example: exampleGenerate,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(genFlag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			}

			genFlag.Name = strings.TrimSuffix(genFlag.Name, ".conf") + ".conf"
			if strings.Trim(genFlag.Output, " ") == "" {
				genFlag.Output = strings.TrimSuffix(genFlag.Name, ".conf") + ".json"
			} else {
				genFlag.Output = strings.TrimSuffix(genFlag.Output, ".json") + ".json"
			}

			initWorkDirectory()
			workPathName := config.GetProfileWorkPath(genFlag.Name)
			homePathName := config.GetProfileHomePath(genFlag.Name)
			_, err := ioutil.ReadFile(workPathName)
			if err != nil {
				_, errinfo := ioutil.ReadFile(homePathName)
				if errinfo != nil {
					fmt.Printf("%s profile.Generate failed, msg: Convert file: %v, read file :%v err:%v\n", ColorString("red", "[ERROR]"), genFlag.Name, homePathName, errinfo)
					os.Exit(1)
				}
			}

			// Determine whether json file already exists
			ParamPath := config.GetGenerateWorkPath(genFlag.Output)
			_, err = os.Stat(ParamPath)
			if err == nil {
				fmt.Printf("%s %s", ColorString("yellow", "[Warning]"), fmt.Sprintf(outputTips, "generated parameter"))
				genFlag.Force = confirm()
				if !genFlag.Force {
					fmt.Printf("outputFile exist and you have given up to overwrite it\n")
					os.Exit(1)
				}
				RunGenerateRemote(genFlag)
			} else {
				RunGenerateRemote(genFlag)
			}

			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&genFlag.Name, "name", "n", "", "profile name, query by command \"keentune profile list\"")
	flags.StringVarP(&genFlag.Output, "output", "o", "", "output parameter configuration file name, default with suffix \".json\"")

	return cmd
}


