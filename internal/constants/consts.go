package constants

const UserDeviceCreationEventType = "com.dimo.zone.device.create"

const (
	SmartCarVendor = "SmartCar"
	TeslaVendor    = "Tesla"
)

const (
	AutoPiVendor      = "AutoPi"
	AutoPiWebhookPath = "/webhooks/autopi-command"
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
