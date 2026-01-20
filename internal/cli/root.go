package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turtacn/Aeterna/internal/monitor"
	"github.com/turtacn/Aeterna/internal/orchestrator"
	"github.com/turtacn/Aeterna/pkg/logger"
	"github.com/turtacn/Aeterna/pkg/protocol"
	"gopkg.in/yaml.v3"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "aeterna",
	Short: "Aeterna: The UPHR-O Process Orchestrator",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the orchestrator daemon",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Load Config
		data, err := os.ReadFile(cfgFile)
		if err != nil {
			fmt.Printf("Error reading config: %v\n", err)
			os.Exit(1)
		}
		var cfg protocol.Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			fmt.Printf("Error parsing config: %v\n", err)
			os.Exit(1)
		}

		// 2. Init Logger & Metrics
		logger.InitLogger(cfg.Observability.LogLevel)
		monitor.InitMetrics(cfg.Observability.MetricsPort)

		logger.Log.Info("Booting Aeterna UPHR-O Engine...", "service", cfg.Service.Name)

		// 3. Start Engine
		engine := orchestrator.NewEngine(&cfg)
		if err := engine.Start(); err != nil {
			logger.Log.Error("Engine fatal error", "err", err)
			os.Exit(1)
		}
	},
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Trigger a hot reload (SIGHUP)",
	Run: func(cmd *cobra.Command, args []string) {
		// In a real CLI, this would find the PID file and send signal
		fmt.Println("Please send SIGHUP to the running Aeterna process (PID 1 in container).")
		// e.g., pkill -HUP aeterna
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "aeterna.yaml", "config file path")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(reloadCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Personal.AI order the ending
