package main

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	ketchv1 "github.com/shipa-corp/ketch/internal/api/v1beta1"
)

const appUpdateHelp = `
Update an app.
`

func newAppUpdateCmd(cfg config, out io.Writer) *cobra.Command {
	options := appUpdateOptions{}
	cmd := &cobra.Command{
		Use:   "update APPNAME",
		Short: "Update an app",
		Long:  appUpdateHelp,
		Args:  cobra.ExactValidArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			options.descriptionSet = cmd.Flags().Changed("description")
			options.dockerRegistrySecretSet = cmd.Flags().Changed("registry-secret")
			options.envsSet = cmd.Flags().Changed("env")
			return appUpdate(cmd.Context(), cfg, options, out)
		},
	}
	cmd.Flags().StringVarP(&options.description, "description", "d", "", "App description")
	cmd.Flags().StringVarP(&options.dockerRegistrySecret, "registry-secret", "", "", "A name of a Secret with docker credentials. This secret must be created in the same namespace of the pool.")
	cmd.MarkFlagRequired("pool")

	return cmd
}

type appUpdateOptions struct {
	name                    string
	descriptionSet          bool
	description             string
	envsSet                 bool
	envs                    []string
	dockerRegistrySecretSet bool
	dockerRegistrySecret    string
}

func appUpdate(ctx context.Context, cfg config, options appUpdateOptions, out io.Writer) error {
	envs, err := getEnvs(options.envs)
	if err != nil {
		return fmt.Errorf("failed to parse env variables: %w", err)
	}
	app := ketchv1.App{}
	if err = cfg.Client().Get(ctx, types.NamespacedName{Name: options.name}, &app); err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}
	if options.descriptionSet {
		app.Spec.Description = options.description
	}
	if options.dockerRegistrySecretSet {
		app.Spec.DockerRegistry = ketchv1.DockerRegistrySpec{
			SecretName: options.dockerRegistrySecret,
		}
	}
	if options.envsSet {
		app.Spec.Env = envs
	}
	if err = cfg.Client().Update(ctx, &app); err != nil {
		return fmt.Errorf("failed to create an app: %w", err)
	}
	fmt.Fprintln(out, "Successfully updated!")
	return nil
}
