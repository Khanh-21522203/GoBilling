package product

import "time"

type Product struct {
	ID          string
	Name        string
	Description *string
	Active      bool
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Product) IsActive() bool {
	return p.Active
}

func (p *Product) Archive() {
	p.Active = false
	p.UpdatedAt = time.Now().UTC()
}

type PricingType string

const (
	PricingTypeFlat   PricingType = "flat"
	PricingTypeTiered PricingType = "tiered"
)

type BillingInterval string

const (
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalYearly  BillingInterval = "yearly"
)

type PricingTier struct {
	UpTo       *int64 `json:"up_to"`
	UnitAmount int64  `json:"unit_amount"`
	FlatAmount int64  `json:"flat_amount"`
}

type Plan struct {
	ID                   string
	ProductID            string
	Name                 string
	Description          *string
	PricingType          PricingType
	Amount               int64
	Currency             string
	BillingInterval      BillingInterval
	BillingIntervalCount int
	TrialPeriodDays      int
	Tiers                []PricingTier
	Active               bool
	Metadata             map[string]string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (p *Plan) IsActive() bool {
	return p.Active
}

func (p *Plan) Archive() {
	p.Active = false
	p.UpdatedAt = time.Now().UTC()
}

func (p *Plan) ValidatePricing() error {
	if p.PricingType == PricingTypeFlat {
		if p.Amount < 0 {
			return ErrInvalidAmount
		}
	} else if p.PricingType == PricingTypeTiered {
		if len(p.Tiers) == 0 {
			return ErrMissingTiers
		}
		lastTier := p.Tiers[len(p.Tiers)-1]
		if lastTier.UpTo != nil {
			return ErrInvalidTiers
		}
	}
	return nil
}
