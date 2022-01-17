package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/gabihodoroaga/cloudrun-casbin/db"
)

// policyUpdateWatcher is the implementation of the casbin.WatcherEx interface
// in order to sync policies between instances
type policyUpdateWatcher struct {
	reload        func(string)
	notifyChannel string
}

// newPolicyUpdateWatcher returns a new policyUpdateWatcher
func newPolicyUpdateWatcher(conn *pgx.Conn, notifyChannel string) *policyUpdateWatcher {
	return &policyUpdateWatcher{
		notifyChannel: notifyChannel,
	}
}

// SetUpdateCallback - is triggered when the watcher is set and will provide
// you a func that can be called to reload the policies in order to respond
// to external events
func (w *policyUpdateWatcher) SetUpdateCallback(f func(string)) error {
	w.reload = f
	return nil
}

// Update is called by the current instance of the enforcer when an update
// ocurred, added or removed policy
func (w *policyUpdateWatcher) Update() error {
	return nil
}

func (w *policyUpdateWatcher) Close() {
	// not implemented
}

// UpdateForAddPolicy is called when a new policy is added to the current
// enforcer instance
func (w *policyUpdateWatcher) UpdateForAddPolicy(sec, ptype string, params ...string) error {
	zap.L().Sugar().Debugf("received UpdateForAddPolicy: %s,%s,%v\n", sec, ptype, params)
	// only grouping policy is used for now
	if sec == "g" && ptype == "g" {
		_, err := db.GetDB().Exec(context.Background(),
			fmt.Sprintf("NOTIFY %s, '%s,%s'", w.notifyChannel, "add", strings.Join(params, ",")))
		if err != nil {
			zap.L().Error(fmt.Sprintf("error exec notify with params %v", params), zap.Error(err))
		}
	}
	return nil
}

// UpdateForRemovePolicy is called when a policy is removed from the current
// enforcer instance
func (w *policyUpdateWatcher) UpdateForRemovePolicy(sec, ptype string, params ...string) error {
	zap.L().Sugar().Debugf("received UpdateForRemovePolicy: %s,%s,%v\n", sec, ptype, params)
	if sec == "g" && ptype == "g" {
		_, err := db.GetDB().Exec(context.Background(),
			fmt.Sprintf("NOTIFY %s, '%s,%s'", w.notifyChannel, "remove", strings.Join(params, ",")))
		if err != nil {
			zap.L().Error(fmt.Sprintf("error exec notify with params %v", params), zap.Error(err))
		}
	}
	return nil
}

func (w *policyUpdateWatcher) UpdateForRemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	// not implemented
	return nil
}

func (w *policyUpdateWatcher) UpdateForSavePolicy(model model.Model) error {
	// not implemented
	return nil
}

func (w *policyUpdateWatcher) UpdateForAddPolicies(sec string, ptype string, rules ...[]string) error {
	// not implemented
	return nil
}

func (w *policyUpdateWatcher) UpdateForRemovePolicies(sec string, ptype string, rules ...[]string) error {
	// not implemented
	return nil
}
