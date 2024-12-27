package models

type Category struct {
	ID          string      `bson:"_id" json:"id"`
	Name        string      `bson:"name" json:"name"`
	Description string      `bson:"description" json:"description"`
	Path        string      `bson:"path" json:"path"`
	ParentID    string      `bson:"parent_id" json:"parentId"`
	Children    []*Category `bson:"-" json:"children"`
}

type CategoryRequest struct {
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Path        string `bson:"path" json:"path" validation:"omitempty"`
	ParentID    string `bson:"parent_id" json:"parentId"`
}

type CategoryRequestMulti struct {
	Categories []CategoryRequest `json:"categories"`
}
