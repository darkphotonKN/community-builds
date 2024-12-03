package item

type CreateItemRequest struct {
	Category string `json:"category" binding:"required,category" db:"category"`
	Class    string `json:"class" binding:"required,class" db:"class"`
	Type     string `json:"type" binding:"required,type" db:"type"`
	Name     string `json:"name" binding:"required,min=2" db:"name"`
	ImageURL string `json:"imageUrl,omitempty" db:"image_url"`
}

type UpdateItemReq struct {
	Category string `json:"category" binding:"required,category" db:"category"`
	Class    string `json:"class" binding:"required,class" db:"class"`
	Type     string `json:"type" binding:"required,type" db:"type"`
	Name     string `json:"name" binding:"required,min=2" db:"name"`
	ImageURL string `json:"imageUrl,omitempty" db:"image_url"`
}

type WikiItem struct {
	ImageUrl    string `json:"imageUrl" db:"image_url"`
	Category    string `json:"category" db:"category"`
	Class       string `json:"class" db:"class"`
	Name        string `json:"name" db:"name"`
	Type        string `json:"type" db:"type"`
	Description string `json:"description" db:"description"`
	// armor
	RequiredLevel        string `json:"requiredLevel" db:"required_level"`
	RequiredStrength     string `json:"requiredStrength" db:"required_strength"`
	RequiredDexterity    string `json:"requiredDexterity" db:"required_dexterity"`
	RequiredIntelligence string `json:"requiredIntelligence" db:"required_intelligence"`
	Armour               string `json:"armour" db:"armour"`
	EnergyShield         string `json:"energyShield" db:"energy_shield"`
	Evasion              string `json:"evasion" db:"evasion"`
	Block                string `json:"block" db:"block"`
	Ward                 string `json:"ward" db:"ward"`
	// weapon
	Damage string `json:"damage" db:"damage"`
	APS    string `json:"aps" db:"aps"`
	Crit   string `json:"crit" db:"crit"`
	PDPS   string `json:"pdps" db:"pdps"`
	EDPS   string `json:"edps" db:"edps"`
	DPS    string `json:"dps" db:"dps"`
	// poison
	Life     string `json:"life" db:"life"`
	Mana     string `json:"mana" db:"mana"`
	Duration string `json:"duration" db:"duration"`
	Usage    string `json:"usage" db:"usage"`
	Capacity string `json:"capacity" db:"capacity"`
	// common
	Additional string   `json:"additional" db:"additional"`
	Stats      []string `json:"stats" db:"stats"`
}

type BaseItem struct {
	ImageUrl string `json:"imageUrl" db:"image_url"`
	Category string `json:"category" db:"category"`
	Class    string `json:"class" db:"class"`
	Name     string `json:"name" db:"name"`
	Type     string `json:"type" db:"type"`

	RequiredLevel        string `json:"requiredLevel" db:"required_level"`
	RequiredStrength     string `json:"requiredStrength" db:"required_strength"`
	RequiredDexterity    string `json:"requiredDexterity" db:"required_dexterity"`
	RequiredIntelligence string `json:"requiredIntelligence" db:"required_intelligence"`

	Damage string `json:"damage" db:"damage"`
	APS    string `json:"aps" db:"aps"`
	Crit   string `json:"crit" db:"crit"`
	DPS    string `json:"dps" db:"dps"`

	Armour       string `json:"armour" db:"armour"`
	Evasion      string `json:"evasion" db:"evasion"`
	EnergyShield string `json:"energyShield" db:"energy_shield"`
	Ward         string `json:"ward" db:"ward"`

	Stats []string `json:"stats" db:"stats"`
}

type ModItem struct {
	Affix string `json:"affix" db:"affix"`
	Name  string `json:"name" db:"name"`
	Level string `json:"level" db:"level"`
	Stat  string `json:"stat" db:"stat"`
	Tags  string `json:"tags" db:"tags"`
}
