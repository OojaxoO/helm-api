package models

import (
	"github.com/jinzhu/gorm"
)

type Cluster struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Config 	     string `json:"config"`
}

// GetArticle Get a single article based on ID
func GetCluster(id int) (*Cluster, error) {
	var cluster Cluster 
	err := db.Where("id = ?", id).First(&cluster).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &cluster, nil
}
