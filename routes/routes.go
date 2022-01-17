package routes

import (
	"net/http"
	"os"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/gabihodoroaga/cloudrun-casbin/db"
	"github.com/gabihodoroaga/cloudrun-casbin/auth"
)

func SetupRoutes(r *gin.Engine) {

	r.Use(auth.RequireAuthentication())
	r.Use(auth.RequireAuthorization())

	r.GET("/api/v1/users/info", func(c *gin.Context) {
		email := c.GetString("user")
		roles, err := auth.GetEnforcer().GetRolesForUser(email)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"email": email,
			"roles": roles,
		})
	})

	r.GET("/api/v1/users", func(c *gin.Context) {
		var users []struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		err := pgxscan.Select(c.Request.Context(), db.GetDB(), &users, "SELECT email, role FROM users_roles;")
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.Wrapf(err, "error get users"))
			return
		}
		hostname, _ := os.Hostname()
		c.Header("X-Server", hostname)
		c.JSON(http.StatusOK, users)
	})

	r.POST("/api/v1/users/:userid/:role", func(c *gin.Context) {
		userid := c.Param("userid")
		role := c.Param("role")
		if userid == "" || role == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid values provided for userid or role"))
			return
		}

		_, err := db.GetDB().Exec(c.Request.Context(), "INSERT INTO users_roles (email, role) VALUES ($1,$2);", userid, role)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.Wrapf(err, "error insert users"))
			return
		}
		auth.AddUserRole(userid, role)
		c.Status(http.StatusOK)
	})


	r.DELETE("/api/v1/users/:userid/:role", func(c *gin.Context) {
		userid := c.Param("userid")
		role := c.Param("role")
		if userid == "" || role == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid values provided for userid or role"))
			return
		}
		_, err := db.GetDB().Exec(c.Request.Context(), "DELETE FROM users_roles WHERE email = $1 AND role = $2;", userid, role)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.Wrapf(err, "error insert users"))
			return
		}
		auth.RemoveUserRole(userid, role)
		c.Status(http.StatusOK)
	})

	r.GET("/api/v1/ping", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.Header("X-Server", hostname)
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/api/v1/ping", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.Header("X-Server", hostname)
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/api/v1/ping/:id", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.Header("X-Server", hostname)
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"id":      c.Param("id"),
		})
	})

	r.DELETE("/api/v1/ping/:id", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.Header("X-Server", hostname)
		c.Status(http.StatusOK)
	})

}