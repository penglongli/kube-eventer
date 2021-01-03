package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/penglongli/kube-eventer/pkg/eventer"
	"github.com/penglongli/kube-eventer/pkg/kube"
	"github.com/penglongli/kube-eventer/pkg/signals"
)

func NewCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "kube-eventer",
		Short: "kube-eventer is used to watch the events of Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			// init kube client
			if err := kube.InitKubeClient(); err != nil {
				panic(err)
			}

			systemSignal := signals.SetupSignalHandler()
			// start kube-eventer
			kubeEventer := eventer.NewKubeEventer(systemSignal)
			if err := kubeEventer.Run(); err != nil {
				klog.Warning("Shutdown.")
			}
		},
	}
	return rootCmd
}
