package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/paulrose/hatch/internal/daemon"
)

var runCmd = &cobra.Command{
	Use:    "_run",
	Short:  "Run the daemon (internal, used by launchd)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGTERM, syscall.SIGINT)
		defer stop()

		d := daemon.New()
		if err := d.Run(ctx); err != nil {
			log.Error().Err(err).Msg("daemon exited with error")
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
