package eventer

import (
	"context"
	"encoding/json"
	"sync"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	"github.com/penglongli/kube-eventer/pkg/election"
	"github.com/penglongli/kube-eventer/pkg/kube"
)

type KubeEventer struct {
	systemSignal <- chan struct{}

	stopCh chan struct{}

	namespaces sync.Map

	wg *sync.WaitGroup
}

func NewKubeEventer(systemSignal <-chan struct{}) *KubeEventer {
	return &KubeEventer{
		systemSignal: systemSignal,
		stopCh: make(chan struct{}),
		wg: &sync.WaitGroup{},
	}
}

func (eventer *KubeEventer) Run() error {
	ctx, cancel := context.WithCancel(context.TODO())

	leaderElection := election.NewLeaderElectionWithConfigMap(kube.GetClientSet(),
		"kube-eventer",
		func(ctx context.Context) {
			if err := eventer.Start(); err != nil {
				klog.Error(err)
			}
			cancel()
		},
	)
	if err := leaderElection.Run(ctx); err != nil {
		return err
	}
	return nil
}

func (eventer *KubeEventer) Start() error {
	namespaceWatch, err := kube.GetClientSet().CoreV1().Namespaces().Watch(context.Background(), v1.ListOptions{
		Watch: true,
		ResourceVersion: "0",
		TimeoutSeconds: &yearSeconds,
	})
	if err != nil {
		klog.Error(err)
		return err
	}

	namespaceChan := namespaceWatch.ResultChan()

LOOP:
	for {
		select {
		case event, ok := <-namespaceChan:
			{
				if !ok {
					break LOOP
				}

				if event.Type == watch.Error {
					continue
				}
				if event.Type == watch.Added || event.Type == watch.Deleted {
					bs, err := json.Marshal(event.Object)
					if err != nil {
						klog.Error(err)
						continue
					}

					namespace := new(corev1.Namespace)
					if err = json.Unmarshal(bs, namespace); err != nil {
						klog.Error(err)
						continue
					}

					if event.Type == watch.Added {
						eventer.handleNamespaceAdd(namespace.Name)
						continue
					}

					if event.Type == watch.Deleted {
						eventer.handleNamespaceDelete(namespace.Name)
						continue
					}
				}
			}
		case <-eventer.stopCh:
			break LOOP
		case <-eventer.systemSignal:
			break LOOP
		}
	}

	eventer.wg.Wait()
	klog.Infof("kube-eventer is closed.")
	return nil
}

func (eventer *KubeEventer) Stop() {
	eventer.namespaces.Range(func(k, v interface{}) bool {
		ns := v.(*KubeNamespace)
		ns.Stop()
		return true
	})
}

func (eventer *KubeEventer) handleNamespaceAdd(namespace string) {
	kubeNamespace := NewKubeNamespace(namespace)
	eventer.namespaces.Store(namespace, kubeNamespace)

	go func() {
		eventer.wg.Add(1)
		_ = kubeNamespace.Watch()
		eventer.wg.Done()
		eventer.namespaces.Delete(namespace)
	}()
}

func (eventer *KubeEventer) handleNamespaceDelete(namespace string) {
	v, ok := eventer.namespaces.Load(namespace)
	if !ok {
		return
	}

	v.(*KubeNamespace).Stop()
	eventer.namespaces.Delete(namespace)
}
