package reconcile

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"
)

type ReconcilerRegistry interface {
	Register(r Reconciler)
	GetReconciler(name string) Reconciler
	UnRegister(name string)
}

// Reconciler is a long running go routine which reconciles to monitor/maintain the state of cluster./**
type Reconciler interface {
	Stop() <-chan struct{}
	Name() string
	Reconcile(ctx context.Context)
}

type ReconcileManager struct {
	ctx           context.Context
	cancel        context.CancelFunc
	reconcilerMap map[string]Reconciler
}

func NewReconcileManager() (*ReconcileManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &ReconcileManager{
		ctx:           ctx,
		cancel:        cancel,
		reconcilerMap: make(map[string]Reconciler),
	}, nil
}

func (rm *ReconcileManager) Register(r Reconciler) {
	rm.reconcilerMap[r.Name()] = r
	r.Reconcile(rm.ctx)
}

func (rm *ReconcileManager) GetReconciler(name string) Reconciler {
	logger := log.From(rm.ctx)
	r, ok := rm.reconcilerMap[name]
	if !ok {
		logger.V(1).Info("Reconciler not found", "reconcilerName", name)
		return nil
	}

	return r
}

func (rm *ReconcileManager) UnRegister(name string) {
	r := rm.GetReconciler(name)
	if r != nil {
		logger := log.From(rm.ctx)
		logger.V(1).Info("Unregistering reconciler", "reconcilerName", r.Name())
		stopped := r.Stop()
		<-stopped

		delete(rm.reconcilerMap, r.Name())
	}
}
