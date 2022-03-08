// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// DeviceIntegration is an object representing the database table.
type DeviceIntegration struct {
	DeviceDefinitionID string    `boil:"device_definition_id" json:"device_definition_id" toml:"device_definition_id" yaml:"device_definition_id"`
	IntegrationID      string    `boil:"integration_id" json:"integration_id" toml:"integration_id" yaml:"integration_id"`
	Country            string    `boil:"country" json:"country" toml:"country" yaml:"country"`
	Capabilities       null.JSON `boil:"capabilities" json:"capabilities,omitempty" toml:"capabilities" yaml:"capabilities,omitempty"`
	CreatedAt          time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt          time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *deviceIntegrationR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L deviceIntegrationL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeviceIntegrationColumns = struct {
	DeviceDefinitionID string
	IntegrationID      string
	Country            string
	Capabilities       string
	CreatedAt          string
	UpdatedAt          string
}{
	DeviceDefinitionID: "device_definition_id",
	IntegrationID:      "integration_id",
	Country:            "country",
	Capabilities:       "capabilities",
	CreatedAt:          "created_at",
	UpdatedAt:          "updated_at",
}

var DeviceIntegrationTableColumns = struct {
	DeviceDefinitionID string
	IntegrationID      string
	Country            string
	Capabilities       string
	CreatedAt          string
	UpdatedAt          string
}{
	DeviceDefinitionID: "device_integrations.device_definition_id",
	IntegrationID:      "device_integrations.integration_id",
	Country:            "device_integrations.country",
	Capabilities:       "device_integrations.capabilities",
	CreatedAt:          "device_integrations.created_at",
	UpdatedAt:          "device_integrations.updated_at",
}

// Generated where

var DeviceIntegrationWhere = struct {
	DeviceDefinitionID whereHelperstring
	IntegrationID      whereHelperstring
	Country            whereHelperstring
	Capabilities       whereHelpernull_JSON
	CreatedAt          whereHelpertime_Time
	UpdatedAt          whereHelpertime_Time
}{
	DeviceDefinitionID: whereHelperstring{field: "\"devices_api\".\"device_integrations\".\"device_definition_id\""},
	IntegrationID:      whereHelperstring{field: "\"devices_api\".\"device_integrations\".\"integration_id\""},
	Country:            whereHelperstring{field: "\"devices_api\".\"device_integrations\".\"country\""},
	Capabilities:       whereHelpernull_JSON{field: "\"devices_api\".\"device_integrations\".\"capabilities\""},
	CreatedAt:          whereHelpertime_Time{field: "\"devices_api\".\"device_integrations\".\"created_at\""},
	UpdatedAt:          whereHelpertime_Time{field: "\"devices_api\".\"device_integrations\".\"updated_at\""},
}

// DeviceIntegrationRels is where relationship names are stored.
var DeviceIntegrationRels = struct {
	DeviceDefinition string
	Integration      string
}{
	DeviceDefinition: "DeviceDefinition",
	Integration:      "Integration",
}

// deviceIntegrationR is where relationships are stored.
type deviceIntegrationR struct {
	DeviceDefinition *DeviceDefinition `boil:"DeviceDefinition" json:"DeviceDefinition" toml:"DeviceDefinition" yaml:"DeviceDefinition"`
	Integration      *Integration      `boil:"Integration" json:"Integration" toml:"Integration" yaml:"Integration"`
}

// NewStruct creates a new relationship struct
func (*deviceIntegrationR) NewStruct() *deviceIntegrationR {
	return &deviceIntegrationR{}
}

// deviceIntegrationL is where Load methods for each relationship are stored.
type deviceIntegrationL struct{}

var (
	deviceIntegrationAllColumns            = []string{"device_definition_id", "integration_id", "country", "capabilities", "created_at", "updated_at"}
	deviceIntegrationColumnsWithoutDefault = []string{"device_definition_id", "integration_id", "country"}
	deviceIntegrationColumnsWithDefault    = []string{"capabilities", "created_at", "updated_at"}
	deviceIntegrationPrimaryKeyColumns     = []string{"device_definition_id", "integration_id", "country"}
	deviceIntegrationGeneratedColumns      = []string{}
)

type (
	// DeviceIntegrationSlice is an alias for a slice of pointers to DeviceIntegration.
	// This should almost always be used instead of []DeviceIntegration.
	DeviceIntegrationSlice []*DeviceIntegration
	// DeviceIntegrationHook is the signature for custom DeviceIntegration hook methods
	DeviceIntegrationHook func(context.Context, boil.ContextExecutor, *DeviceIntegration) error

	deviceIntegrationQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deviceIntegrationType                 = reflect.TypeOf(&DeviceIntegration{})
	deviceIntegrationMapping              = queries.MakeStructMapping(deviceIntegrationType)
	deviceIntegrationPrimaryKeyMapping, _ = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, deviceIntegrationPrimaryKeyColumns)
	deviceIntegrationInsertCacheMut       sync.RWMutex
	deviceIntegrationInsertCache          = make(map[string]insertCache)
	deviceIntegrationUpdateCacheMut       sync.RWMutex
	deviceIntegrationUpdateCache          = make(map[string]updateCache)
	deviceIntegrationUpsertCacheMut       sync.RWMutex
	deviceIntegrationUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deviceIntegrationAfterSelectHooks []DeviceIntegrationHook

var deviceIntegrationBeforeInsertHooks []DeviceIntegrationHook
var deviceIntegrationAfterInsertHooks []DeviceIntegrationHook

var deviceIntegrationBeforeUpdateHooks []DeviceIntegrationHook
var deviceIntegrationAfterUpdateHooks []DeviceIntegrationHook

var deviceIntegrationBeforeDeleteHooks []DeviceIntegrationHook
var deviceIntegrationAfterDeleteHooks []DeviceIntegrationHook

var deviceIntegrationBeforeUpsertHooks []DeviceIntegrationHook
var deviceIntegrationAfterUpsertHooks []DeviceIntegrationHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DeviceIntegration) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DeviceIntegration) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DeviceIntegration) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DeviceIntegration) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DeviceIntegration) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DeviceIntegration) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DeviceIntegration) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DeviceIntegration) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DeviceIntegration) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range deviceIntegrationAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeviceIntegrationHook registers your hook function for all future operations.
func AddDeviceIntegrationHook(hookPoint boil.HookPoint, deviceIntegrationHook DeviceIntegrationHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		deviceIntegrationAfterSelectHooks = append(deviceIntegrationAfterSelectHooks, deviceIntegrationHook)
	case boil.BeforeInsertHook:
		deviceIntegrationBeforeInsertHooks = append(deviceIntegrationBeforeInsertHooks, deviceIntegrationHook)
	case boil.AfterInsertHook:
		deviceIntegrationAfterInsertHooks = append(deviceIntegrationAfterInsertHooks, deviceIntegrationHook)
	case boil.BeforeUpdateHook:
		deviceIntegrationBeforeUpdateHooks = append(deviceIntegrationBeforeUpdateHooks, deviceIntegrationHook)
	case boil.AfterUpdateHook:
		deviceIntegrationAfterUpdateHooks = append(deviceIntegrationAfterUpdateHooks, deviceIntegrationHook)
	case boil.BeforeDeleteHook:
		deviceIntegrationBeforeDeleteHooks = append(deviceIntegrationBeforeDeleteHooks, deviceIntegrationHook)
	case boil.AfterDeleteHook:
		deviceIntegrationAfterDeleteHooks = append(deviceIntegrationAfterDeleteHooks, deviceIntegrationHook)
	case boil.BeforeUpsertHook:
		deviceIntegrationBeforeUpsertHooks = append(deviceIntegrationBeforeUpsertHooks, deviceIntegrationHook)
	case boil.AfterUpsertHook:
		deviceIntegrationAfterUpsertHooks = append(deviceIntegrationAfterUpsertHooks, deviceIntegrationHook)
	}
}

// One returns a single deviceIntegration record from the query.
func (q deviceIntegrationQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DeviceIntegration, error) {
	o := &DeviceIntegration{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for device_integrations")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DeviceIntegration records from the query.
func (q deviceIntegrationQuery) All(ctx context.Context, exec boil.ContextExecutor) (DeviceIntegrationSlice, error) {
	var o []*DeviceIntegration

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to DeviceIntegration slice")
	}

	if len(deviceIntegrationAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DeviceIntegration records in the query.
func (q deviceIntegrationQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count device_integrations rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deviceIntegrationQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if device_integrations exists")
	}

	return count > 0, nil
}

// DeviceDefinition pointed to by the foreign key.
func (o *DeviceIntegration) DeviceDefinition(mods ...qm.QueryMod) deviceDefinitionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.DeviceDefinitionID),
	}

	queryMods = append(queryMods, mods...)

	query := DeviceDefinitions(queryMods...)
	queries.SetFrom(query.Query, "\"devices_api\".\"device_definitions\"")

	return query
}

// Integration pointed to by the foreign key.
func (o *DeviceIntegration) Integration(mods ...qm.QueryMod) integrationQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.IntegrationID),
	}

	queryMods = append(queryMods, mods...)

	query := Integrations(queryMods...)
	queries.SetFrom(query.Query, "\"devices_api\".\"integrations\"")

	return query
}

// LoadDeviceDefinition allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (deviceIntegrationL) LoadDeviceDefinition(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDeviceIntegration interface{}, mods queries.Applicator) error {
	var slice []*DeviceIntegration
	var object *DeviceIntegration

	if singular {
		object = maybeDeviceIntegration.(*DeviceIntegration)
	} else {
		slice = *maybeDeviceIntegration.(*[]*DeviceIntegration)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &deviceIntegrationR{}
		}
		args = append(args, object.DeviceDefinitionID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceIntegrationR{}
			}

			for _, a := range args {
				if a == obj.DeviceDefinitionID {
					continue Outer
				}
			}

			args = append(args, obj.DeviceDefinitionID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`devices_api.device_definitions`),
		qm.WhereIn(`devices_api.device_definitions.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load DeviceDefinition")
	}

	var resultSlice []*DeviceDefinition
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice DeviceDefinition")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for device_definitions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for device_definitions")
	}

	if len(deviceIntegrationAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.DeviceDefinition = foreign
		if foreign.R == nil {
			foreign.R = &deviceDefinitionR{}
		}
		foreign.R.DeviceIntegrations = append(foreign.R.DeviceIntegrations, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.DeviceDefinitionID == foreign.ID {
				local.R.DeviceDefinition = foreign
				if foreign.R == nil {
					foreign.R = &deviceDefinitionR{}
				}
				foreign.R.DeviceIntegrations = append(foreign.R.DeviceIntegrations, local)
				break
			}
		}
	}

	return nil
}

// LoadIntegration allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (deviceIntegrationL) LoadIntegration(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDeviceIntegration interface{}, mods queries.Applicator) error {
	var slice []*DeviceIntegration
	var object *DeviceIntegration

	if singular {
		object = maybeDeviceIntegration.(*DeviceIntegration)
	} else {
		slice = *maybeDeviceIntegration.(*[]*DeviceIntegration)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &deviceIntegrationR{}
		}
		args = append(args, object.IntegrationID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceIntegrationR{}
			}

			for _, a := range args {
				if a == obj.IntegrationID {
					continue Outer
				}
			}

			args = append(args, obj.IntegrationID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`devices_api.integrations`),
		qm.WhereIn(`devices_api.integrations.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Integration")
	}

	var resultSlice []*Integration
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Integration")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for integrations")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for integrations")
	}

	if len(deviceIntegrationAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Integration = foreign
		if foreign.R == nil {
			foreign.R = &integrationR{}
		}
		foreign.R.DeviceIntegrations = append(foreign.R.DeviceIntegrations, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.IntegrationID == foreign.ID {
				local.R.Integration = foreign
				if foreign.R == nil {
					foreign.R = &integrationR{}
				}
				foreign.R.DeviceIntegrations = append(foreign.R.DeviceIntegrations, local)
				break
			}
		}
	}

	return nil
}

// SetDeviceDefinition of the deviceIntegration to the related item.
// Sets o.R.DeviceDefinition to related.
// Adds o to related.R.DeviceIntegrations.
func (o *DeviceIntegration) SetDeviceDefinition(ctx context.Context, exec boil.ContextExecutor, insert bool, related *DeviceDefinition) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"device_integrations\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"device_definition_id"}),
		strmangle.WhereClause("\"", "\"", 2, deviceIntegrationPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.DeviceDefinitionID, o.IntegrationID, o.Country}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.DeviceDefinitionID = related.ID
	if o.R == nil {
		o.R = &deviceIntegrationR{
			DeviceDefinition: related,
		}
	} else {
		o.R.DeviceDefinition = related
	}

	if related.R == nil {
		related.R = &deviceDefinitionR{
			DeviceIntegrations: DeviceIntegrationSlice{o},
		}
	} else {
		related.R.DeviceIntegrations = append(related.R.DeviceIntegrations, o)
	}

	return nil
}

// SetIntegration of the deviceIntegration to the related item.
// Sets o.R.Integration to related.
// Adds o to related.R.DeviceIntegrations.
func (o *DeviceIntegration) SetIntegration(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Integration) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"device_integrations\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"integration_id"}),
		strmangle.WhereClause("\"", "\"", 2, deviceIntegrationPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.DeviceDefinitionID, o.IntegrationID, o.Country}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.IntegrationID = related.ID
	if o.R == nil {
		o.R = &deviceIntegrationR{
			Integration: related,
		}
	} else {
		o.R.Integration = related
	}

	if related.R == nil {
		related.R = &integrationR{
			DeviceIntegrations: DeviceIntegrationSlice{o},
		}
	} else {
		related.R.DeviceIntegrations = append(related.R.DeviceIntegrations, o)
	}

	return nil
}

// DeviceIntegrations retrieves all the records using an executor.
func DeviceIntegrations(mods ...qm.QueryMod) deviceIntegrationQuery {
	mods = append(mods, qm.From("\"devices_api\".\"device_integrations\""))
	return deviceIntegrationQuery{NewQuery(mods...)}
}

// FindDeviceIntegration retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDeviceIntegration(ctx context.Context, exec boil.ContextExecutor, deviceDefinitionID string, integrationID string, country string, selectCols ...string) (*DeviceIntegration, error) {
	deviceIntegrationObj := &DeviceIntegration{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"device_integrations\" where \"device_definition_id\"=$1 AND \"integration_id\"=$2 AND \"country\"=$3", sel,
	)

	q := queries.Raw(query, deviceDefinitionID, integrationID, country)

	err := q.Bind(ctx, exec, deviceIntegrationObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from device_integrations")
	}

	if err = deviceIntegrationObj.doAfterSelectHooks(ctx, exec); err != nil {
		return deviceIntegrationObj, err
	}

	return deviceIntegrationObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DeviceIntegration) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_integrations provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = currTime
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(deviceIntegrationColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deviceIntegrationInsertCacheMut.RLock()
	cache, cached := deviceIntegrationInsertCache[key]
	deviceIntegrationInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deviceIntegrationAllColumns,
			deviceIntegrationColumnsWithDefault,
			deviceIntegrationColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"device_integrations\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"device_integrations\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into device_integrations")
	}

	if !cached {
		deviceIntegrationInsertCacheMut.Lock()
		deviceIntegrationInsertCache[key] = cache
		deviceIntegrationInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DeviceIntegration.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DeviceIntegration) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deviceIntegrationUpdateCacheMut.RLock()
	cache, cached := deviceIntegrationUpdateCache[key]
	deviceIntegrationUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deviceIntegrationAllColumns,
			deviceIntegrationPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update device_integrations, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"device_integrations\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, deviceIntegrationPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, append(wl, deviceIntegrationPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update device_integrations row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for device_integrations")
	}

	if !cached {
		deviceIntegrationUpdateCacheMut.Lock()
		deviceIntegrationUpdateCache[key] = cache
		deviceIntegrationUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deviceIntegrationQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for device_integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for device_integrations")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeviceIntegrationSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceIntegrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"device_integrations\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, deviceIntegrationPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in deviceIntegration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all deviceIntegration")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DeviceIntegration) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no device_integrations provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if o.CreatedAt.IsZero() {
			o.CreatedAt = currTime
		}
		o.UpdatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(deviceIntegrationColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	deviceIntegrationUpsertCacheMut.RLock()
	cache, cached := deviceIntegrationUpsertCache[key]
	deviceIntegrationUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deviceIntegrationAllColumns,
			deviceIntegrationColumnsWithDefault,
			deviceIntegrationColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			deviceIntegrationAllColumns,
			deviceIntegrationPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert device_integrations, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(deviceIntegrationPrimaryKeyColumns))
			copy(conflict, deviceIntegrationPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"device_integrations\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deviceIntegrationType, deviceIntegrationMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert device_integrations")
	}

	if !cached {
		deviceIntegrationUpsertCacheMut.Lock()
		deviceIntegrationUpsertCache[key] = cache
		deviceIntegrationUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DeviceIntegration record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DeviceIntegration) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no DeviceIntegration provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), deviceIntegrationPrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"device_integrations\" WHERE \"device_definition_id\"=$1 AND \"integration_id\"=$2 AND \"country\"=$3"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from device_integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for device_integrations")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deviceIntegrationQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no deviceIntegrationQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from device_integrations")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_integrations")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeviceIntegrationSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deviceIntegrationBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceIntegrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"device_integrations\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceIntegrationPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from deviceIntegration slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for device_integrations")
	}

	if len(deviceIntegrationAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *DeviceIntegration) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDeviceIntegration(ctx, exec, o.DeviceDefinitionID, o.IntegrationID, o.Country)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeviceIntegrationSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeviceIntegrationSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deviceIntegrationPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"device_integrations\".* FROM \"devices_api\".\"device_integrations\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deviceIntegrationPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in DeviceIntegrationSlice")
	}

	*o = slice

	return nil
}

// DeviceIntegrationExists checks if the DeviceIntegration row exists.
func DeviceIntegrationExists(ctx context.Context, exec boil.ContextExecutor, deviceDefinitionID string, integrationID string, country string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"device_integrations\" where \"device_definition_id\"=$1 AND \"integration_id\"=$2 AND \"country\"=$3 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, deviceDefinitionID, integrationID, country)
	}
	row := exec.QueryRowContext(ctx, sql, deviceDefinitionID, integrationID, country)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if device_integrations exists")
	}

	return exists, nil
}
