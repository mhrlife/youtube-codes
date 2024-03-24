package controller

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/labstack/echo/v4"
	"time"
)

func (ec *EchoController) tagUrls() {
	g := ec.e.Group("/tag")
	g.POST("/", ec.createTag)
	g.GET("/", ec.listTags)
	g.GET("/:slug/", ec.listTagArticles)
}

type createTagRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (ec *EchoController) createTag(c echo.Context) error {
	return ec.endpointMetric.Do("create_tag", func() error {
		request, err := Bind[createTagRequest](c)
		if err != nil {
			return err
		}

		// write requests can not be canceled by the client
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		tag := entity.Tag{
			Slug: request.Slug,
			Name: request.Name,
		}
		if err := ec.db.CreateTag(ctx, &tag); err != nil {
			_ = c.String(500, err.Error())
			return err
		}

		return c.JSON(201, tag)
	})
}

func (ec *EchoController) listTags(c echo.Context) error {
	return ec.endpointMetric.Do("list_tags", func() error {
		tags, err := ec.db.ListTags(c.Request().Context())
		if err != nil {
			return err
		}
		return c.JSON(200, tags)
	})
}

func (ec *EchoController) listTagArticles(c echo.Context) error {
	return ec.endpointMetric.Do("list_tag_articles", func() error {
		articles, err := ec.db.ListTagArticles(c.Request().Context(), c.Param("slug"))
		if err != nil {
			return err
		}
		return c.JSON(200, articles)
	})
}
