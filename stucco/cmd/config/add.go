package configcmd

import (
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func addCommand() *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "add",
		Short: "add resolver/scalar/etc. to stucco.json/.yaml",
	}
	configCommand.AddCommand(resolverCommand())
	configCommand.AddCommand(interfaceCommand())
	configCommand.AddCommand(unionCommand())
	configCommand.AddCommand(scalarCommand())
	configCommand.AddCommand(schemaCommand())
	return configCommand
}

func schemaCommand() *cobra.Command {
	addScalarCommand := &cobra.Command{
		Use:   "schema",
		Short: "Add schema to stucco.json/.yaml [arg1: string]",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := addToConfig()
			if err != nil {
				klog.Fatalln(err.Error())
			}

			cfg.AddSchema(args[0])

			err = utils.SaveConfigFile("stucco", cfg)

			if err != nil {
				klog.Fatalln(err.Error())
			}
		},
	}
	return addScalarCommand
}

func scalarCommand() *cobra.Command {
	addScalarCommand := &cobra.Command{
		Use:   "scalar",
		Args:  cobra.MinimumNArgs(2),
		Short: "Add scalar to stucco.json/.yaml [arg1: Name, arg2: Parse, arg3: Serialize]",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := addToConfig()
			if err != nil {
				klog.Fatalln(err.Error())
			}

			if len(args) == 2 {
				args = append(args, "")
			}

			cfg.AddScalar(args[0], args[1], args[2])

			err = utils.SaveConfigFile("stucco", cfg)

			if err != nil {
				klog.Fatalln(err.Error())
			}
		},
	}
	return addScalarCommand
}

func unionCommand() *cobra.Command {
	addUnionCommand := &cobra.Command{
		Use:   "union",
		Short: "Add union to stucco.json/.yaml",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := addToConfig()
			if err != nil {
				klog.Fatalln(err.Error())
			}

			cfg.AddUnion(args[0], args[1])

			err = utils.SaveConfigFile("stucco", cfg)

			if err != nil {
				klog.Fatalln(err.Error())
			}
		},
	}
	return addUnionCommand
}

func interfaceCommand() *cobra.Command {
	addInterfaceCommand := &cobra.Command{
		Use:   "interface",
		Short: "Add interface to stucco.json/.yaml",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := addToConfig()
			if err != nil {
				klog.Fatalln(err.Error())
			}

			cfg.AddInterface(args[0], args[1])

			err = utils.SaveConfigFile("stucco", cfg)

			if err != nil {
				klog.Fatalln(err.Error())
			}
		},
	}
	return addInterfaceCommand
}

func resolverCommand() *cobra.Command {
	addResolverCommand := &cobra.Command{
		Use:   "resolver",
		Short: "Add resolver to stucco.json/.yaml",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := addToConfig()
			if err != nil {
				klog.Fatalln(err.Error())
			}

			cfg.AddResolver(args[0], args[1])

			err = utils.SaveConfigFile("stucco", cfg)

			if err != nil {
				klog.Fatalln(err.Error())
			}
		},
	}
	return addResolverCommand
}
