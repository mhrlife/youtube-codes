package entity

import "gorm.io/gorm"

type Author struct {
	gorm.Model

	DisplayName string `json:"display_name"`
}

type Tag struct {
	gorm.Model

	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Article struct {
	gorm.Model

	Title   string `json:"title"`
	Content string `json:"content"`

	AuthorID uint   `json:"author_id"`
	Author   Author `json:"author"`

	Tags []Tag `gorm:"many2many:article_tags;"`
}
