package controller

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	MicomManager  = "MICOM_MANAGER"
	DeviceBringup = "DEVICE_BRINGUP"
	Navigation    = "NAVIGATION"
	Middleware    = "MIDDLEWARE"
)

type Repo struct {
	Id          int
	Repo_name   string
	Repo_path   string
	Ver_major   int
	Ver_minor_1 int
	Ver_minor_2 int
	Updated_at  string
}

type DBController struct {
	db *gorm.DB
}

func NewDB(path *string) *DBController {
	db, err := gorm.Open(sqlite.Open(*path), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	db.AutoMigrate(&Repo{})
	return &DBController{
		db: db,
	}
}

func (c *DBController) InsertRepo(updateType *string, repo *Repo) {
	c.db.Table(*updateType).Create(repo)
}

func (c *DBController) UpdateRepo(updateType *string, repo *Repo) {
	c.db.Table(*updateType).Save(repo)
}

func (c *DBController) DeleteRepo(updateType *string, repo *Repo) {
	c.db.Table(*updateType).Delete(repo)
}

func (c *DBController) GetRepo(updateType *string, repo *Repo) {
	c.db.Table(*updateType).First(repo, repo.Id)
}

func (c *DBController) GetAllRepos(updateType *string, repos *[]Repo) {
	c.db.Table(*updateType).Find(repos)
}

func (c *DBController) Close() {
	sqlDB, err := c.db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	sqlDB.Close()
}
