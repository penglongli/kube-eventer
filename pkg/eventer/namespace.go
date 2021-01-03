package eventer

import (
	"context"
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/penglongli/kube-eventer/pkg/kube"
)

type KubeNamespace struct {
	quitCh chan struct{}
	Name string
}

func NewKubeNamespace(name string) *KubeNamespace {
	return &KubeNamespace{
		quitCh: make(chan struct{}),
		Name: name,
	}
}

func (kn *KubeNamespace) Watch() error {
	ch, err := kn.getChannel()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			klog.Error(r)
		}
	}()

	resultChan := ch.ResultChan()
LOOP:
	for {
		select {
		case event, ok := <-resultChan:
			{
				if !ok {
					break LOOP
				}

				if event.Type == watch.Error {
					continue
				}

				if event.Type == watch.Added {
					bs, err := json.Marshal(event.Object)
					if err != nil {
						klog.Error(err)
						continue
					}

					event := new(corev1.Event)
					if err = json.Unmarshal(bs, event); err != nil {
						klog.Error(err)
						continue
					}
					//TODO
				}
			}
		case <-kn.quitCh:
			break LOOP
		}
	}
	if ch != nil {
		ch.Stop()
	}

	klog.Infof("%s closed.", kn.Name)
	return nil
}

func (kn *KubeNamespace) Stop() {
	close(kn.quitCh)
}

func (kn *KubeNamespace) getChannel() (ch watch.Interface, err error) {
	ch, err = kube.GetClientSet().CoreV1().Events(kn.Name).Watch(context.Background(), v1.ListOptions{
		Watch: true,
		ResourceVersion: "0",
		TimeoutSeconds: &yearSeconds,
	})
	if err != nil {
		klog.Error(err)
	}
	return ch, err
}
