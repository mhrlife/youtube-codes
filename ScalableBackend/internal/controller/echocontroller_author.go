package controller

import (
	"ScalableBackend/internal/database"
	"ScalableBackend/internal/entity"
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"strconv"
	"time"
)

func (ec *EchoController) authorUrls() {
	g := ec.e.Group("/author")
	g.POST("/", ec.createAuthor)
	g.GET("/:id", ec.getAuthor)
}

type createAuthorRequest struct {
	DisplayName string `json:"display_name"`
}

func (ec *EchoController) createAuthor(c echo.Context) error {
	return ec.endpointMetric.Do("create_author", func() error {
		request, err := Bind[createAuthorRequest](c)
		if err != nil {
			return err
		}

		// write requests can not be canceled by the client
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		author := entity.Author{
			DisplayName: request.DisplayName,
		}
		if err := ec.db.CreateAuthor(ctx, &author); err != nil {
			_ = c.String(500, err.Error())
			return err
		}

		return c.JSON(201, author)
	})
}

func (ec *EchoController) getAuthor(c echo.Context) error {
	return ec.endpointMetric.Do("get_author", func() error {
		authorId, err := strconv.Atoi(c.Param("id"))
		if err != nil { //todo: handle these kind of errors with 401 error
			return err
		}

		author, err := ec.db.GetAuthor(c.Request().Context(), uint(authorId))
		if err != nil {
			if errors.Is(err, database.ErrEntityNotFound) {
				return c.String(404, "user not found")
			}
			return err
		}

		return c.JSON(200, author)
	})
}
