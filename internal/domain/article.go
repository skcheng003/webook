package domain

import "time"

type Article struct {
	Id         int64
	Title      string
	Content    string
	Author     Author
	Status     ArticleStatus
	CreateTime time.Time
	UpdateTime time.Time
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

const (
	ArticleStatusUnknown = iota
	ArticleStatusPrivate
	ArticleStatusUnPublished
	ArticleStatusPublished
)
