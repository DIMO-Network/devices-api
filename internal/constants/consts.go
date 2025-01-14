package constants

const UserDeviceCreationEventType = "com.dimo.zone.device.create"

const (
	PowerTrainTypeKey = "powertrain_type"
)

const (
	SmartCarVendor = "SmartCar"
	TeslaVendor    = "Tesla"
)

const (
	AutoPiVendor      = "AutoPi"
)

const (
	IntegrationStyleAddon   string = "Addon"
	IntegrationStyleOEM     string = "OEM"
	IntegrationStyleWebhook string = "Webhook"
)

const (
	IntegrationTypeHardware string = "Hardware"
	IntegrationTypeAPI      string = "API"
)

const (
	TeslaAPIV1 int = 1
	TeslaAPIV2 int = 2
)

// AutoPiSubStatusEnum integration sub-status
type AutoPiSubStatusEnum string

const (
	PendingSoftwareUpdate      AutoPiSubStatusEnum = "PendingSoftwareUpdate"
	CompletedSoftwareUpdate    AutoPiSubStatusEnum = "CompletedSoftwareUpdate"
	QueriedDeviceOk            AutoPiSubStatusEnum = "QueriedDeviceOk"
	PatchedVehicleProfile      AutoPiSubStatusEnum = "PatchedVehicleProfile"
	AssociatedDeviceToTemplate AutoPiSubStatusEnum = "AssociatedDeviceToTemplate"
	AppliedTemplate            AutoPiSubStatusEnum = "AppliedTemplate"
	PendingTemplateConfirm     AutoPiSubStatusEnum = "PendingTemplateConfirm"
	TemplateConfirmed          AutoPiSubStatusEnum = "TemplateConfirmed"
)

func (r AutoPiSubStatusEnum) String() string {
	return string(r)
}

const (
	ChargeLimit        string = "charge/limit"
	FrunkOpen          string = "frunk/open"
	TrunkOpen          string = "trunk/open"
	DoorsLock          string = "doors/lock"
	DoorsUnlock        string = "doors/unlock"
	TelemetrySubscribe string = "telemetry/subscribe"
)
