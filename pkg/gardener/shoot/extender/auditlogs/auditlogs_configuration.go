package auditlogs

import "fmt"

var (
	ErrConfigurationNotFound = fmt.Errorf("audit logs configuration not found")
)

type region = string

type providerType = string

type AuditLogData struct {
	TenantID   string `json:"tenantID" validate:"required"`
	ServiceURL string `json:"serviceURL" validate:"required,url"`
	SecretName string `json:"secretName" validate:"required"`
}

type Configuration map[providerType]map[region]AuditLogData

func (a Configuration) GetAuditLogData(providerType, region string) (AuditLogData, error) {
	providerCfg, found := (a)[providerType]
	if !found {
		return AuditLogData{}, fmt.Errorf("%w: missing providerType: '%s'",
			ErrConfigurationNotFound,
			providerType)
	}

	providerCfgForRegion, found := providerCfg[region]
	if !found {
		return AuditLogData{}, fmt.Errorf("%w: missing region: '%s' for providerType: '%s'",
			ErrConfigurationNotFound,
			region,
			providerType)
	}

	return providerCfgForRegion, nil
}
