package auth

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SetupAuth loads the casbin permission model
func SetupAuth() error {
	ctx := context.TODO()

	enf, err := casbin.NewEnforcer("auth/casbin_model.conf", "auth/casbin_permissions.csv")
	if err != nil {
		return errors.Wrap(err, "unable to create the casbin enforcer")
	}
	enforcer = enf

	zap.L().Info("auth.init: loaded casbin model and permission")

	err = reloadGroupPolicy(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to read group policies")
	}

	err = setupNotificationsListener(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to setup the policy update notifier")
	}

	zap.L().Info("auth.init: loaded users and roles")
	return nil
}
