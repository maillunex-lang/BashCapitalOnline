package model

type Labor struct {
	Worker   int `json:"worker"`
	Engineer int `json:"engineer"`
	Manager  int `json:"manager"`
}

type ItemDef struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Tier     string `json:"tier"`
	Category string `json:"category"`
}

type RecipeIO struct {
	ItemID string `json:"item_id"`
	Qty    int    `json:"qty"`
}

type RecipeDef struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	BuildingType       string   `json:"building_type"`
	Inputs             []RecipeIO `json:"inputs"`
	Output             RecipeIO `json:"output"`
	Labor              Labor    `json:"labor"`
	Energy             int      `json:"energy"`
	ResearchCost       int      `json:"research_cost"`
	SeasonBonusMonths  []int    `json:"season_bonus_months"`
}

type BuildingDef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	BuildCost int    `json:"build_cost"`
	Zone      string `json:"zone"`
	Footprint int    `json:"footprint"`
}

type CityDef struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	X               float64  `json:"x"`
	Y               float64  `json:"y"`
	Focus           string   `json:"focus"`
	StartPopulation int      `json:"start_population"`
	StartRadius     int      `json:"start_radius"`
	Resources       []string `json:"resources"`
	Labor           Labor    `json:"labor"`
	Enabled         bool     `json:"enabled"`
}

type RouteDef struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Ticks       int     `json:"ticks"`
	CostPerUnit float64 `json:"cost_per_unit"`
	Mode        string  `json:"mode"`
}

type Catalog struct {
	Items     []ItemDef
	Recipes   []RecipeDef
	Buildings []BuildingDef
	Cities    []CityDef
	Routes    []RouteDef
	ItemByID  map[string]ItemDef
	RecipeByID map[string]RecipeDef
	BuildingByID map[string]BuildingDef
	CityByID  map[string]CityDef
}

type Parcel struct {
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Zone     string  `json:"zone"` // center, residential, commercial, industrial, rural, empty
	OwnerID  string  `json:"owner_id,omitempty"`
	Building string  `json:"building_id,omitempty"` // instance id
	Slots    int     `json:"retail_slots"`
	RentedBy string  `json:"rented_by,omitempty"`
	Traffic  float64 `json:"traffic"`
	HasRoad  bool    `json:"has_road"`
}

type BuildingInst struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	FirmID       string         `json:"firm_id"`
	CityID       string         `json:"city_id"`
	X            int            `json:"x"`
	Y            int            `json:"y"`
	RecipeID     string         `json:"recipe_id"`
	RetoolWeeks  int            `json:"retool_weeks"`
	Inventory    map[string]int `json:"inventory"`
	Quality      float64        `json:"quality"`
	Active       bool           `json:"active"`
}

type ShopLease struct {
	ID           string  `json:"id"`
	FirmID       string  `json:"firm_id"`
	CityID       string  `json:"city_id"`
	X            int     `json:"x"`
	Y            int     `json:"y"`
	ItemID       string  `json:"item_id"`
	Price        float64 `json:"price"`
	Stock        int     `json:"stock"`
	FixedRent    float64 `json:"fixed_rent"`
	RevenueShare float64 `json:"revenue_share"`
	Radius       int     `json:"radius"` // зона охвата в клетках (манхэттен)
}

// DeputyTask — задание для ИИ-заместителя (доступно по подписке).
type DeputyTask struct {
	ID       string            `json:"id"`
	Kind     string            `json:"kind"` // procure_inputs|price_retail|restock|sell_excess|research|auto_contracts
	Label    string            `json:"label"`
	Params   map[string]string `json:"params,omitempty"`
	Active   bool              `json:"active"`
	LastNote string            `json:"last_note,omitempty"`
}

type Deputy struct {
	Hired bool         `json:"hired"`
	Name  string       `json:"name"`
	Tasks []DeputyTask `json:"tasks"`
}

type Firm struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	IsAI               bool               `json:"is_ai"`
	Industry           string             `json:"industry"`
	Cash               float64            `json:"cash"`
	Researched         map[string]bool    `json:"researched"`
	BrandAwareness     map[string]float64 `json:"brand_awareness"`
	Trucks             int                `json:"trucks"`
	Credit             float64            `json:"credit"`
	CapitalSnapshot    float64            `json:"capital_snapshot"`
	SubscriptionActive bool               `json:"subscription_active"`
	SubscriptionUntil  int                `json:"subscription_until_tick"` // world tick expiry
	Deputy             *Deputy            `json:"deputy,omitempty"`
}

type MarketOrder struct {
	ID       string  `json:"id"`
	FirmID   string  `json:"firm_id"`
	CityID   string  `json:"city_id"`
	ItemID   string  `json:"item_id"`
	Side     string  `json:"side"` // buy|sell
	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	Quality  float64 `json:"quality"`
	Filled   int     `json:"filled"`
}

type Contract struct {
	ID         string  `json:"id"`
	SellerID   string  `json:"seller_id"`
	BuyerID    string  `json:"buyer_id"`
	ItemID     string  `json:"item_id"`
	CityFrom   string  `json:"city_from"`
	CityTo     string  `json:"city_to"`
	QtyPerTick int     `json:"qty_per_tick"`
	Price      float64 `json:"price"`
	Quality    float64 `json:"quality"`
	TicksLeft  int     `json:"ticks_left"`
	UseOwnFleet bool   `json:"use_own_fleet"`
}

type Shipment struct {
	ID         string  `json:"id"`
	FirmID     string  `json:"firm_id"`
	ItemID     string  `json:"item_id"`
	Qty        int     `json:"qty"`
	Quality    float64 `json:"quality"`
	FromCity   string  `json:"from_city"`
	ToCity     string  `json:"to_city"`
	TicksLeft  int     `json:"ticks_left"`
	CostPaid   float64 `json:"cost_paid"`
	OwnFleet   bool    `json:"own_fleet"`
}

// ResourceDeposit — конечная залежь на карте; после исчерпания появляется в другом месте.
type ResourceDeposit struct {
	ID        string  `json:"id"`
	ItemID    string  `json:"item_id"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Amount    int     `json:"amount"`
	MaxAmount int     `json:"max_amount"`
	Quality   float64 `json:"quality"`
}

type CityState struct {
	ID           string             `json:"id"`
	Population   int                `json:"population"`
	Radius       int                `json:"radius"`
	LaborPool    Labor              `json:"labor_pool"`
	LaborUsed    Labor              `json:"labor_used"`
	Parcels      []Parcel           `json:"parcels"`
	Deposits     []ResourceDeposit  `json:"deposits"`
	DemandSeg    map[string]float64 `json:"demand_segments"` // category -> unmet desire units
	HousingOcc   float64            `json:"housing_occupancy"`
}

type World struct {
	Tick           int                    `json:"tick"`
	Month          int                    `json:"month"` // 1-12
	Year           int                    `json:"year"`
	Cities         map[string]*CityState  `json:"cities"`
	Firms          map[string]*Firm       `json:"firms"`
	Buildings      map[string]*BuildingInst `json:"buildings"`
	Shops          map[string]*ShopLease  `json:"shops"`
	Orders         map[string]*MarketOrder `json:"orders"`
	Contracts      map[string]*Contract   `json:"contracts"`
	Shipments      map[string]*Shipment   `json:"shipments"`
	Events         []string               `json:"events"`
	LastTickLog    []string               `json:"last_tick_log"`
}
