package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string
	Pets []Pet
}

// TopPetOwner is a view, which returns owners with the most pets.
// the view will be named as top_pet_owner.
type TopPetOwner struct{}

// ViewDef returns the view definition
func (TopPetOwner) ViewDef(db *gorm.DB) gorm.ViewOption {
	return gorm.ViewOption{
		Replace: true,
		Query: db.
			Table("users").
			Select("users.id, users.name, count(pets.id) as pet_count").
			Joins("left join pets on pets.user_id = users.id").
			Group("users.id").
			Order("pet_count desc").
			Limit(10),
	}
}

type ViewDefiner interface {
	ViewDef(db *gorm.DB) gorm.ViewOption
}
