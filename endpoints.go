package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func endpoints(e *echo.Echo, m HistoryManager) {

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST},
	}))

	e.POST("/visit", func(c echo.Context) error {
		d := uLoc{}
		if err := c.Bind(&d); err != nil {
			fmt.Println(err.Error())
			return c.NoContent(http.StatusBadRequest)
		}
		id, err := m.WriteHistory(c.Request().Context(), d)
		if err != nil {
			fmt.Println(err.Error())
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, externalID(id))
	})

	e.GET("/visit", func(c echo.Context) error {
		vID := c.QueryParam("visitId")

		if len(vID) > 0 {
			h, err := m.GetHistoryByVisitID(c.Request().Context(), vID)
			if err != nil {
				fmt.Println(err.Error())
				return c.NoContent(http.StatusInternalServerError)
			}

			return c.JSON(http.StatusOK, h)
		}

		userID := c.QueryParam("userId")
		searchString := c.QueryParam("searchString")

		if len(userID) == 0 || len(searchString) == 0 {
			return c.JSON(http.StatusBadRequest, `{"message":"query requires either visitId or both userId and searchString"}`)
		}

		h, err := m.GetHistoryByUserID(c.Request().Context(), userID, searchString)
		if err != nil {
			fmt.Println(err.Error())
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, h)
	})
}
