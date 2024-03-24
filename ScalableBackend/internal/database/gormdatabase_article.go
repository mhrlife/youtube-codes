package database

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/sirupsen/logrus"
)

func (g *GormDatabase) CreateTag(ctx context.Context, tag *entity.Tag) error {
	return g.queryMetric.Do("create_tag", func() error {
		err := g.db.WithContext(ctx).Create(tag).Error
		if err != nil {
			logrus.WithError(err).Error("error while creating the user")
			return err
		}
		return nil
	})
}

func (g *GormDatabase) ListTags(ctx context.Context) ([]entity.Tag, error) {
	var tags []entity.Tag
	return tags, g.queryMetric.Do("list_tags", func() error {
		err := g.db.WithContext(ctx).Find(&tags).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (g *GormDatabase) ListTagArticles(ctx context.Context, slug string) ([]entity.Article, error) {
	var articles []entity.Article
	return articles, g.queryMetric.Do("list_tag_articles", func() error {
		err := g.db.WithContext(ctx).Joins("JOIN article_tags on article_tags.article_id = articles.id").
			Joins("JOIN tags on tags.id = article_tags.tag_id").
			Joins("JOIN authors on authors.id = articles.author_id").
			Where("tags.slug = ?", slug).
			Find(&articles).Error
		if err != nil {
			logrus.WithError(err).WithField("slug", slug).Error("couldn't load tag articles")
			return err
		}

		return nil
	})
}
