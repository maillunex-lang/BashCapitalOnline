package model

type Labor struct {
	Worker   int `json:"worker"`
	Engineer int `json:"engineer"`
	Manager  int `json:"manager"`
}

type ItemDef struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Tier       string `json:"tier"`
	Category   string `json:"category"`
	Branch     string `json:"branch,omitempty"`
	TierUnlock int    `json:"tier_unlock,omitempty"`
	Icon       string `json:"icon,omitempty"`
}

type RecipeIO struct {
	ItemID string `json:"item_id"`
	Qty    int    `json:"qty"`
}

type RecipeDef struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	BuildingType      string     `json:"building_type"`
	Inputs            []RecipeIO `json:"inputs"`
	Output            RecipeIO   `json:"output"`
	Labor             Labor      `json:"labor"`
	Energy            int        `json:"energy"`
	ResearchCost      int        `json:"research_cost"`
	SeasonBonusMonths []int      `json:"season_bonus_months"`
	Branch            string     `json:"branch,omitempty"`
	TierUnlock        int        `json:"tier_unlock,omitempty"`
}

type BuildingDef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"` // extract|factory|logistics
	BuildCost int    `json:"build_cost"`
	Zone      string `json:"zone"`
	Footprint int    `json:"footprint"`
	// IndustryTag — специализация как в Capitalism Lab (food/furniture/steel…).
	IndustryTag string `json:"industry_tag,omitempty"`
	Branch      string `json:"branch,omitempty"`
	TierUnlock  int    `json:"tier_unlock,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// ShopDef — тип розничной точки: продуктовый не продаёт электронику и т.п.
type ShopDef struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	RentMult           float64  `json:"rent_mult"`
	AllowedCategories  []string `json:"allowed_categories"`
	AllowedTiers       []string `json:"allowed_tiers"`
	AllowedItems       []string `json:"allowed_items,omitempty"` // если задано — белый список
	Icon               string   `json:"icon,omitempty"`
}

// InputSource — откуда завод закупает вход (свой склад / город / биржа / конкретная фирма).
type InputSource struct {
	Mode   string `json:"mode"`              // own|city|market|firm
	FirmID string `json:"firm_id,omitempty"` // для mode=firm
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

type PriceCorridor struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type CityMarketOffer struct {
	ItemID       string  `json:"item_id"`
	Price        float64 `json:"price"`
	Quality      float64 `json:"quality"`
	StockPerTick int     `json:"stock_per_tick"`
}

type ChatConfig struct {
	MaxMessages           int `json:"max_messages"`
	AIMessagesPerTickCap  int `json:"ai_messages_per_tick_cap"`
	MinTicksBetweenAIMsg  int `json:"min_ticks_between_ai_msg"`
}

// LogisticsConfig — доставка по расстоянию (как в жизни): тики = f(дистанция), минимум 1.
type LogisticsConfig struct {
	UnitsPerTick     float64 `json:"units_per_tick"`      // map units per 1 tick of travel
	CostPerUnitDist  float64 `json:"cost_per_unit_dist"`  // ₽ per map-unit per cargo unit
	LocalCostPerUnit float64 `json:"local_cost_per_unit"` // same-city last mile
	BaseCostPerUnit  float64 `json:"base_cost_per_unit"`  // fixed loading fee
}

// BalanceConfig — ставки из scripts/rebalance_economy.py (единый источник с коридорами).
type BalanceConfig struct {
	Version            int                `json:"version"`
	WagePerTick        map[string]float64 `json:"wage_per_tick"`
	EnergyCostExtract  float64            `json:"energy_cost_extract"`
	EnergyCostFactory  float64            `json:"energy_cost_factory"`
	TargetMargin       map[string]float64 `json:"target_margin"`
	Notes              string             `json:"notes"`
}

type EconomyConfig struct {
	PriceCorridors map[string]PriceCorridor `json:"price_corridors"`
	CityMarket     []CityMarketOffer        `json:"city_market"`
	Chat           ChatConfig               `json:"chat"`
	Logistics      LogisticsConfig          `json:"logistics"`
	Balance        BalanceConfig            `json:"balance"`
}

type AIPersona struct {
	ID            string  `json:"id"`
	Skill         string  `json:"skill"`
	Focus         string  `json:"focus"`
	Style         string  `json:"style"`
	ChatTone      string  `json:"chat_tone"`
	Aggression    float64 `json:"aggression"`
	NoobMistakes  bool    `json:"noob_mistakes"`
	Spawn         bool    `json:"spawn"`
	Name          string  `json:"name"`
	Industry      string  `json:"industry"`
	City          string  `json:"city"`
	Cash          float64 `json:"cash"`
}

type RnDNode struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Cost            int      `json:"cost"`
	Requires        []string `json:"requires"`
	UnlocksRecipes  []string `json:"unlocks_recipes"`
}

type RnDTree struct {
	Nodes []RnDNode `json:"nodes"`
}

type Catalog struct {
	Items        []ItemDef
	Recipes      []RecipeDef
	Buildings    []BuildingDef
	Shops        []ShopDef
	Cities       []CityDef
	Routes       []RouteDef
	Economy      EconomyConfig
	AIPersonas   []AIPersona
	RnDTree      RnDTree
	ItemByID     map[string]ItemDef
	RecipeByID   map[string]RecipeDef
	BuildingByID map[string]BuildingDef
	ShopByID     map[string]ShopDef
	CityByID     map[string]CityDef
	PersonaByID  map[string]AIPersona
}

type Parcel struct {
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Zone      string  `json:"zone"` // center, residential, commercial, industrial, rural, empty
	OwnerID   string  `json:"owner_id,omitempty"`
	Building  string  `json:"building_id,omitempty"` // instance id
	Slots     int     `json:"retail_slots"`
	RentedBy  string  `json:"rented_by,omitempty"`
	Traffic   float64 `json:"traffic"`
	HasRoad   bool    `json:"has_road"`
	Elevation float64 `json:"elevation"` // mild heightmap for map_3d / iso offset
}

type BuildingInst struct {
	ID           string                   `json:"id"`
	Type         string                   `json:"type"`
	FirmID       string                   `json:"firm_id"`
	CityID       string                   `json:"city_id"`
	X            int                      `json:"x"`
	Y            int                      `json:"y"`
	RecipeID     string                   `json:"recipe_id"`
	RetoolWeeks  int                      `json:"retool_weeks"`
	Inventory    map[string]int           `json:"inventory"`
	Quality      float64                  `json:"quality"`
	Active       bool                     `json:"active"`
	// AutoProcure — перед тиком производства докупать входы по InputSources.
	AutoProcure bool `json:"auto_procure"`
	// DefaultInput — источник по умолчанию для всех входов рецепта.
	DefaultInput InputSource `json:"default_input"`
	// InputSources — переопределение источника по item_id.
	InputSources map[string]InputSource `json:"input_sources,omitempty"`
	// SellExternal=false — завод/склад не продаёт другим фирмам (только себе).
	SellExternal bool `json:"sell_external"`
}

type ShopLease struct {
	ID           string  `json:"id"`
	FirmID       string  `json:"firm_id"`
	CityID       string  `json:"city_id"`
	X            int     `json:"x"`
	Y            int     `json:"y"`
	ShopType     string  `json:"shop_type"` // grocery|furniture_shop|…
	ItemID       string  `json:"item_id"`
	Price        float64 `json:"price"`
	Stock        int     `json:"stock"`
	FixedRent    float64 `json:"fixed_rent"`
	RevenueShare float64 `json:"revenue_share"`
	Radius       int     `json:"radius"` // зона охвата в клетках (манхэттен)
	// StockSource — откуда пополнять витрину: own|city|market
	StockSource string `json:"stock_source"`
	AutoRestock bool   `json:"auto_restock"`
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
	PersonaID          string             `json:"persona_id,omitempty"`
	LastChatTick       int                `json:"last_chat_tick,omitempty"`
	AccountID          string             `json:"account_id,omitempty"`
	// RPG research progression
	PrimaryBranch string          `json:"primary_branch,omitempty"`
	ResearchNodes map[string]bool `json:"research_nodes,omitempty"`
	ResearchPoints float64        `json:"research_points,omitempty"`
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
	Distance   float64 `json:"distance"` // map units (≈km scale)
	Mode       string  `json:"mode"`     // local|road|rail
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

type ChatMessage struct {
	ID        string `json:"id"`
	Tick      int    `json:"tick"`
	FirmID    string `json:"firm_id,omitempty"`
	Author    string `json:"author"`
	Channel   string `json:"channel"` // global|city:ufa|system
	Text      string `json:"text"`
	IsAI      bool   `json:"is_ai"`
	CreatedAt int64  `json:"created_at"`
}

type CityMarketStock struct {
	ItemID  string  `json:"item_id"`
	Price   float64 `json:"price"`
	Quality float64 `json:"quality"`
	Stock   int     `json:"stock"`
}

type World struct {
	Tick           int                      `json:"tick"`
	Month          int                      `json:"month"` // 1-12
	Year           int                      `json:"year"`
	Cities         map[string]*CityState    `json:"cities"`
	Firms          map[string]*Firm         `json:"firms"`
	Buildings      map[string]*BuildingInst `json:"buildings"`
	Shops          map[string]*ShopLease    `json:"shops"`
	Orders         map[string]*MarketOrder  `json:"orders"`
	Contracts      map[string]*Contract     `json:"contracts"`
	Shipments      map[string]*Shipment     `json:"shipments"`
	Events         []string                 `json:"events"`
	LastTickLog    []string                 `json:"last_tick_log"`
	Chat           []ChatMessage            `json:"chat"`
	CityMarket     map[string]CityMarketStock `json:"city_market"` // item_id -> offer
	MapRevision    int                      `json:"map_revision"`
	DirtyTiles     []DirtyTile              `json:"dirty_tiles,omitempty"`
}

type DirtyTile struct {
	CityID string `json:"city_id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Reason string `json:"reason"`
}
