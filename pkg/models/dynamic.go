package models

type DynamicType string

const (
	FurnitureType             = "furnitures"
	GiftsAndOccasionsType     = "gifts-and-occasions"
	ClothingType              = "clothing"
	ArtAndCollectiblesType    = "art-and-collectibles"
	AceessoriesAndJewelryType = "accessories-and-jewelry"
	HomeAndLivingType         = "home-and-living"
)

type Clothing struct {
	Color     string   `bson:"color" json:"color"`
	Fabric    string   `json:"fabric" bson:"fabric"`
	Size      string   `json:"size" bson:"size"`
	Scale     string   `json:"scale" bson:"scale" validate:"oneof=EU US/CA"`
	Materials []string `json:"materials" bson:"materials"`
}

type Furniture struct {
	Color      string   `bson:"color" json:"color"`
	Material   string   `json:"material" bson:"material"`
	Dimensions string   `json:"dimensions" bson:"dimensions"` // e.g. "80x60x40cm"
	Weight     float64  `json:"weight" bson:"weight"`         // in kilograms
	RoomType   string   `json:"room_type" bson:"room_type"`   // e.g. "Living Room", "Bedroom"
	Style      string   `json:"style" bson:"style"`           // e.g. "Modern", "Vintage"
	Features   []string `json:"features" bson:"features"`     // e.g. ["Foldable", "Storage Included"]
}

type GiftsAndOccasions struct {
	Color        string   `json:"color" bson:"color"`
	Occasion     string   `json:"occasion" bson:"occasion"`         // e.g. "Birthday", "Wedding"
	Recipient    string   `json:"recipient" bson:"recipient"`       // e.g. "For Him", "For Kids"
	Personalized bool     `json:"personalized" bson:"personalized"` // is customization available
	Materials    []string `json:"materials" bson:"materials"`
	Theme        string   `json:"theme" bson:"theme"` // e.g. "Romantic", "Funny"
}

type ArtAndCollectibles struct {
	Color      string      `json:"color" bson:"color"`
	Medium     string      `json:"medium" bson:"medium"`         // e.g. "Oil", "Acrylic", "Watercolor"
	Style      string      `json:"style" bson:"style"`           // e.g. "Abstract", "Realism"
	Dimensions Measurement `json:"dimensions" bson:"dimensions"` // e.g. "50x70cm"
	Original   bool        `json:"original" bson:"original"`     // true if it's an original piece
	Framed     bool        `json:"framed" bson:"framed"`         // true if frame is included
	Materials  []string    `json:"materials" bson:"materials"`
}

type AceessoriesAndJewelry struct {
	Color    string `json:"color" bson:"color"`
	Material string `json:"material" bson:"material"` // e.g. "Gold", "Beads"
	Type     string `json:"type" bson:"type"`         // e.g. "Necklace", "Bracelet"
	Gender   string `json:"gender" bson:"gender"`     // e.g. "Unisex", "Women's"
	Occasion string `json:"occasion" bson:"occasion"` // e.g. "Wedding", "Everyday"
	Size     string `json:"size" bson:"size"`         // e.g. "Adjustable", "Small"
	Style    string `json:"style" bson:"style"`       // e.g. "Bohemian", "Classic"
}

type HomeAndLiving struct {
	Color      string   `json:"color" bson:"color"`
	Material   string   `json:"material" bson:"material"`     // e.g. "Cotton", "Bamboo"
	Type       string   `json:"type" bson:"type"`             // e.g. "Curtain", "Tableware"
	Room       string   `json:"room" bson:"room"`             // e.g. "Kitchen", "Bedroom"
	Style      string   `json:"style" bson:"style"`           // e.g. "Rustic", "Contemporary"
	Dimensions string   `json:"dimensions" bson:"dimensions"` // e.g. "100x200cm"
	Features   []string `json:"features" bson:"features"`     // e.g. ["Washable", "Eco-friendly"]
}
