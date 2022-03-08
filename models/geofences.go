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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// Geofence is an object representing the database table.
type Geofence struct {
	ID        string            `boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID    string            `boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	Name      string            `boil:"name" json:"name" toml:"name" yaml:"name"`
	Type      string            `boil:"type" json:"type" toml:"type" yaml:"type"`
	H3Indexes types.StringArray `boil:"h3_indexes" json:"h3_indexes,omitempty" toml:"h3_indexes" yaml:"h3_indexes,omitempty"`
	CreatedAt time.Time         `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt time.Time         `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *geofenceR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L geofenceL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var GeofenceColumns = struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	H3Indexes string
	CreatedAt string
	UpdatedAt string
}{
	ID:        "id",
	UserID:    "user_id",
	Name:      "name",
	Type:      "type",
	H3Indexes: "h3_indexes",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}

var GeofenceTableColumns = struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	H3Indexes string
	CreatedAt string
	UpdatedAt string
}{
	ID:        "geofences.id",
	UserID:    "geofences.user_id",
	Name:      "geofences.name",
	Type:      "geofences.type",
	H3Indexes: "geofences.h3_indexes",
	CreatedAt: "geofences.created_at",
	UpdatedAt: "geofences.updated_at",
}

// Generated where

type whereHelpertypes_StringArray struct{ field string }

func (w whereHelpertypes_StringArray) EQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpertypes_StringArray) NEQ(x types.StringArray) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpertypes_StringArray) LT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_StringArray) LTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_StringArray) GT(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_StringArray) GTE(x types.StringArray) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpertypes_StringArray) IsNull() qm.QueryMod { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpertypes_StringArray) IsNotNull() qm.QueryMod {
	return qmhelper.WhereIsNotNull(w.field)
}

var GeofenceWhere = struct {
	ID        whereHelperstring
	UserID    whereHelperstring
	Name      whereHelperstring
	Type      whereHelperstring
	H3Indexes whereHelpertypes_StringArray
	CreatedAt whereHelpertime_Time
	UpdatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"devices_api\".\"geofences\".\"id\""},
	UserID:    whereHelperstring{field: "\"devices_api\".\"geofences\".\"user_id\""},
	Name:      whereHelperstring{field: "\"devices_api\".\"geofences\".\"name\""},
	Type:      whereHelperstring{field: "\"devices_api\".\"geofences\".\"type\""},
	H3Indexes: whereHelpertypes_StringArray{field: "\"devices_api\".\"geofences\".\"h3_indexes\""},
	CreatedAt: whereHelpertime_Time{field: "\"devices_api\".\"geofences\".\"created_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"devices_api\".\"geofences\".\"updated_at\""},
}

// GeofenceRels is where relationship names are stored.
var GeofenceRels = struct {
	UserDeviceToGeofences string
}{
	UserDeviceToGeofences: "UserDeviceToGeofences",
}

// geofenceR is where relationships are stored.
type geofenceR struct {
	UserDeviceToGeofences UserDeviceToGeofenceSlice `boil:"UserDeviceToGeofences" json:"UserDeviceToGeofences" toml:"UserDeviceToGeofences" yaml:"UserDeviceToGeofences"`
}

// NewStruct creates a new relationship struct
func (*geofenceR) NewStruct() *geofenceR {
	return &geofenceR{}
}

// geofenceL is where Load methods for each relationship are stored.
type geofenceL struct{}

var (
	geofenceAllColumns            = []string{"id", "user_id", "name", "type", "h3_indexes", "created_at", "updated_at"}
	geofenceColumnsWithoutDefault = []string{"id", "user_id", "name"}
	geofenceColumnsWithDefault    = []string{"type", "h3_indexes", "created_at", "updated_at"}
	geofencePrimaryKeyColumns     = []string{"id"}
	geofenceGeneratedColumns      = []string{}
)

type (
	// GeofenceSlice is an alias for a slice of pointers to Geofence.
	// This should almost always be used instead of []Geofence.
	GeofenceSlice []*Geofence
	// GeofenceHook is the signature for custom Geofence hook methods
	GeofenceHook func(context.Context, boil.ContextExecutor, *Geofence) error

	geofenceQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	geofenceType                 = reflect.TypeOf(&Geofence{})
	geofenceMapping              = queries.MakeStructMapping(geofenceType)
	geofencePrimaryKeyMapping, _ = queries.BindMapping(geofenceType, geofenceMapping, geofencePrimaryKeyColumns)
	geofenceInsertCacheMut       sync.RWMutex
	geofenceInsertCache          = make(map[string]insertCache)
	geofenceUpdateCacheMut       sync.RWMutex
	geofenceUpdateCache          = make(map[string]updateCache)
	geofenceUpsertCacheMut       sync.RWMutex
	geofenceUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var geofenceAfterSelectHooks []GeofenceHook

var geofenceBeforeInsertHooks []GeofenceHook
var geofenceAfterInsertHooks []GeofenceHook

var geofenceBeforeUpdateHooks []GeofenceHook
var geofenceAfterUpdateHooks []GeofenceHook

var geofenceBeforeDeleteHooks []GeofenceHook
var geofenceAfterDeleteHooks []GeofenceHook

var geofenceBeforeUpsertHooks []GeofenceHook
var geofenceAfterUpsertHooks []GeofenceHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Geofence) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Geofence) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Geofence) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Geofence) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Geofence) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Geofence) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Geofence) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Geofence) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Geofence) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range geofenceAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddGeofenceHook registers your hook function for all future operations.
func AddGeofenceHook(hookPoint boil.HookPoint, geofenceHook GeofenceHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		geofenceAfterSelectHooks = append(geofenceAfterSelectHooks, geofenceHook)
	case boil.BeforeInsertHook:
		geofenceBeforeInsertHooks = append(geofenceBeforeInsertHooks, geofenceHook)
	case boil.AfterInsertHook:
		geofenceAfterInsertHooks = append(geofenceAfterInsertHooks, geofenceHook)
	case boil.BeforeUpdateHook:
		geofenceBeforeUpdateHooks = append(geofenceBeforeUpdateHooks, geofenceHook)
	case boil.AfterUpdateHook:
		geofenceAfterUpdateHooks = append(geofenceAfterUpdateHooks, geofenceHook)
	case boil.BeforeDeleteHook:
		geofenceBeforeDeleteHooks = append(geofenceBeforeDeleteHooks, geofenceHook)
	case boil.AfterDeleteHook:
		geofenceAfterDeleteHooks = append(geofenceAfterDeleteHooks, geofenceHook)
	case boil.BeforeUpsertHook:
		geofenceBeforeUpsertHooks = append(geofenceBeforeUpsertHooks, geofenceHook)
	case boil.AfterUpsertHook:
		geofenceAfterUpsertHooks = append(geofenceAfterUpsertHooks, geofenceHook)
	}
}

// One returns a single geofence record from the query.
func (q geofenceQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Geofence, error) {
	o := &Geofence{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for geofences")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Geofence records from the query.
func (q geofenceQuery) All(ctx context.Context, exec boil.ContextExecutor) (GeofenceSlice, error) {
	var o []*Geofence

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Geofence slice")
	}

	if len(geofenceAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Geofence records in the query.
func (q geofenceQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count geofences rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q geofenceQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if geofences exists")
	}

	return count > 0, nil
}

// UserDeviceToGeofences retrieves all the user_device_to_geofence's UserDeviceToGeofences with an executor.
func (o *Geofence) UserDeviceToGeofences(mods ...qm.QueryMod) userDeviceToGeofenceQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"devices_api\".\"user_device_to_geofence\".\"geofence_id\"=?", o.ID),
	)

	query := UserDeviceToGeofences(queryMods...)
	queries.SetFrom(query.Query, "\"devices_api\".\"user_device_to_geofence\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"devices_api\".\"user_device_to_geofence\".*"})
	}

	return query
}

// LoadUserDeviceToGeofences allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (geofenceL) LoadUserDeviceToGeofences(ctx context.Context, e boil.ContextExecutor, singular bool, maybeGeofence interface{}, mods queries.Applicator) error {
	var slice []*Geofence
	var object *Geofence

	if singular {
		object = maybeGeofence.(*Geofence)
	} else {
		slice = *maybeGeofence.(*[]*Geofence)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &geofenceR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &geofenceR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`devices_api.user_device_to_geofence`),
		qm.WhereIn(`devices_api.user_device_to_geofence.geofence_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load user_device_to_geofence")
	}

	var resultSlice []*UserDeviceToGeofence
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice user_device_to_geofence")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on user_device_to_geofence")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user_device_to_geofence")
	}

	if len(userDeviceToGeofenceAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.UserDeviceToGeofences = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &userDeviceToGeofenceR{}
			}
			foreign.R.Geofence = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.GeofenceID {
				local.R.UserDeviceToGeofences = append(local.R.UserDeviceToGeofences, foreign)
				if foreign.R == nil {
					foreign.R = &userDeviceToGeofenceR{}
				}
				foreign.R.Geofence = local
				break
			}
		}
	}

	return nil
}

// AddUserDeviceToGeofences adds the given related objects to the existing relationships
// of the geofence, optionally inserting them as new records.
// Appends related to o.R.UserDeviceToGeofences.
// Sets related.R.Geofence appropriately.
func (o *Geofence) AddUserDeviceToGeofences(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*UserDeviceToGeofence) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.GeofenceID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"devices_api\".\"user_device_to_geofence\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"geofence_id"}),
				strmangle.WhereClause("\"", "\"", 2, userDeviceToGeofencePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.UserDeviceID, rel.GeofenceID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.GeofenceID = o.ID
		}
	}

	if o.R == nil {
		o.R = &geofenceR{
			UserDeviceToGeofences: related,
		}
	} else {
		o.R.UserDeviceToGeofences = append(o.R.UserDeviceToGeofences, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &userDeviceToGeofenceR{
				Geofence: o,
			}
		} else {
			rel.R.Geofence = o
		}
	}
	return nil
}

// Geofences retrieves all the records using an executor.
func Geofences(mods ...qm.QueryMod) geofenceQuery {
	mods = append(mods, qm.From("\"devices_api\".\"geofences\""))
	return geofenceQuery{NewQuery(mods...)}
}

// FindGeofence retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindGeofence(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*Geofence, error) {
	geofenceObj := &Geofence{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"geofences\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, geofenceObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from geofences")
	}

	if err = geofenceObj.doAfterSelectHooks(ctx, exec); err != nil {
		return geofenceObj, err
	}

	return geofenceObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Geofence) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no geofences provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(geofenceColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	geofenceInsertCacheMut.RLock()
	cache, cached := geofenceInsertCache[key]
	geofenceInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			geofenceAllColumns,
			geofenceColumnsWithDefault,
			geofenceColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(geofenceType, geofenceMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(geofenceType, geofenceMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"geofences\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"geofences\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into geofences")
	}

	if !cached {
		geofenceInsertCacheMut.Lock()
		geofenceInsertCache[key] = cache
		geofenceInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Geofence.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Geofence) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	geofenceUpdateCacheMut.RLock()
	cache, cached := geofenceUpdateCache[key]
	geofenceUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			geofenceAllColumns,
			geofencePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update geofences, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"geofences\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, geofencePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(geofenceType, geofenceMapping, append(wl, geofencePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update geofences row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for geofences")
	}

	if !cached {
		geofenceUpdateCacheMut.Lock()
		geofenceUpdateCache[key] = cache
		geofenceUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q geofenceQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for geofences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for geofences")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o GeofenceSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), geofencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"geofences\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, geofencePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in geofence slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all geofence")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Geofence) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no geofences provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(geofenceColumnsWithDefault, o)

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

	geofenceUpsertCacheMut.RLock()
	cache, cached := geofenceUpsertCache[key]
	geofenceUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			geofenceAllColumns,
			geofenceColumnsWithDefault,
			geofenceColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			geofenceAllColumns,
			geofencePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert geofences, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(geofencePrimaryKeyColumns))
			copy(conflict, geofencePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"geofences\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(geofenceType, geofenceMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(geofenceType, geofenceMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert geofences")
	}

	if !cached {
		geofenceUpsertCacheMut.Lock()
		geofenceUpsertCache[key] = cache
		geofenceUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Geofence record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Geofence) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Geofence provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), geofencePrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"geofences\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from geofences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for geofences")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q geofenceQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no geofenceQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from geofences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for geofences")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o GeofenceSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(geofenceBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), geofencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"geofences\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, geofencePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from geofence slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for geofences")
	}

	if len(geofenceAfterDeleteHooks) != 0 {
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
func (o *Geofence) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindGeofence(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *GeofenceSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := GeofenceSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), geofencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"geofences\".* FROM \"devices_api\".\"geofences\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, geofencePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in GeofenceSlice")
	}

	*o = slice

	return nil
}

// GeofenceExists checks if the Geofence row exists.
func GeofenceExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"geofences\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if geofences exists")
	}

	return exists, nil
}
