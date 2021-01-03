package main

import (
	"k8s.io/klog"

	"github.com/penglongli/kube-eventer/cmd"
)

func main() {
	if err := cmd.NewCommand(); err != nil {
		klog.Error(err)
	}
}
