package hooks

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/coapcloud/coap-hooks-router/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Pass this with header like this:
//
// Authorization: Bearer {APIKEY}

type HooksAPIServer struct {
	store Repository
	addr  net.Addr
}

func ListenAndServe(hooksRepo Repository, port int) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal(err)
	}

	s := &HooksAPIServer{
		store: hooksRepo,
		addr:  addr,
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.AddTrailingSlash())

	hooksGroup := e.Group("/api/hooks")
	hooksGroup.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return subtle.ConstantTimeCompare([]byte(key), config.AdminBearer) == 1, nil
	}))

	hooksGroup.POST("/", func(c echo.Context) error {
		var reqHook Hook
		err := c.Bind(&reqHook)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "owner can't be left blank"})
		}

		err = s.store.CreateHook(reqHook)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, reqHook)
	})

	hooksGroup.GET("/:owner", func(c echo.Context) error {
		owner := c.Param("owner")
		if owner == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "owner can't be left blank"})
		}

		hooks, err := s.store.ListHooksForOwner(owner)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, hooks)
	})

	hooksGroup.DELETE("/:owner/:name", func(c echo.Context) error {
		owner := c.Param("owner")
		if owner == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "owner can't be left blank"})
		}

		name := c.Param("name")
		if name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name can't be left blank"})
		}

		err := s.store.DeleteHookByOwnerAndName(owner, name)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"msg": fmt.Sprintf("hook: %s deleted for owner: %s", name, owner)})
	})

	hooksGroup.DELETE("/:owner", func(c echo.Context) error {
		owner := c.Param("owner")
		if owner == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "owner can't be left blank"})
		}

		err := s.store.DeleteHooksForOwner(owner)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]string{"msg": fmt.Sprintf("hooks deleted for owner: %s", owner)})
	})

	// Start server
	e.Logger.Fatal(e.Start(s.addr.String()))
}
