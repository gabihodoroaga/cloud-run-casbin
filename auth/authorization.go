package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/gabihodoroaga/cloudrun-casbin/db"
	"github.com/gabihodoroaga/cloudrun-casbin/config"
)

var (
	enforcer *casbin.Enforcer
	basePath string = "/api/v1"
)

func GetEnforcer() *casbin.Enforcer {
	return enforcer
}

// RequireAuthorization is the gin middles that checks if the use had the 
// required permission to access the path and method
func RequireAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {

		logger := zap.L().With(zap.String("request_id", c.GetString("request_id")))

		sub := c.GetString("user")
		obj := strings.TrimPrefix(c.Request.URL.Path, basePath)
		act := c.Request.Method

		allow, err := enforcer.Enforce(sub, obj, act)
		logger.Sugar().Debugf("check casbin with args: sub=%s, obj=%s, act=%s", sub, obj, act)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.Wrap(err, "error while calling casbin.Enforce"))
			return
		}

		if !allow {
			logger.Debug("enforcer failed with this reasons")
			c.AbortWithStatus(403)
			return
		}
	}
}

func reloadGroupPolicy(ctx context.Context) error {
	row := struct {
		email string
		role  string
	}{}
	query := "SELECT email, role FROM users_roles;"
	_, err := db.GetDB().QueryFunc(ctx, query, []interface{}{}, []interface{}{&row.email, &row.role},
		func(pgx.QueryFuncRow) error {
			_, err := enforcer.AddGroupingPolicy(row.email, row.role)
			return err
		},
	)
	return err
}

func setupNotificationsListener(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, config.GetConfig().ConnString)
	if err != nil {
		return err
	}

	casbinNotifyChannel := "casbin_notify"
	_, err = conn.Exec(ctx, "LISTEN "+casbinNotifyChannel)

	if err != nil {
		return err
	}

	err = enforcer.SetWatcher(newPolicyUpdateWatcher(conn, casbinNotifyChannel))
	if err != nil {
		return err
	}

	go func() {
		for {
			zap.L().Sugar().Debugf("notificationsListener: waiting for notifications on channels %s", casbinNotifyChannel)
			notification, err := conn.WaitForNotification(ctx)
			if err != nil {
				if errors.Is(err, ctx.Err()) {
					zap.L().Info("notificationsListener: context done, exiting...")
					conn.Close(ctx)
					return
				}
				zap.L().Error("notificationsListener: error waiting for notification", zap.Error(err))
			}
			zap.L().Sugar().Debugf("notificationsListener: received notification on channel %s, pid %d, payload %s", notification.Channel, notification.PID, notification.Payload)

			params := strings.Split(notification.Payload, ",")
			if len(params) != 4 {
				zap.L().Sugar().Errorf("notificationsListener: invalid policy update notification received '%s', expected 4 values separated by comma", notification.Payload)
			}
			switch params[0] {
			case "add":
				enforcer.AddGroupingPolicy(params[1:])
			case "remove":
				enforcer.RemoveGroupingPolicy(params[1:])
			default:
				zap.L().Sugar().Errorf("notificationsListener: invalid policy update notification received '%s', first value must be one of 'add' or 'remove'", notification.Payload)
			}
		}
	}()

	return nil
}

// AddUserRole adds a new casbin grouping policy for the user and role 
func AddUserRole(email, role string) {
	_, err := enforcer.AddGroupingPolicy(email, role)
	if err != nil {
		zap.L().Error("error AddUserRole", zap.Error(errors.Wrap(err, fmt.Sprintf("error AddGroupingPolicy, values email: %s, role:%s", email, role))))
	}
}

// RemoveUserRole removes the casbin grouping policy for the user and role and
func RemoveUserRole(email, role  string) {
	_, err := enforcer.RemoveGroupingPolicy(email, role)
	if err != nil {
		zap.L().Error("error AddUserRole", zap.Error(errors.Wrap(err, fmt.Sprintf("error AddGroupingPolicy, values email: %s, role:%s", email, role))))
	}
}
