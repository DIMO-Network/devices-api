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

// UserDeviceDatum is an object representing the database table.
type UserDeviceDatum struct {
	UserDeviceID string    `boil:"user_device_id" json:"user_device_id" toml:"user_device_id" yaml:"user_device_id"`
	Data         null.JSON `boil:"data" json:"data,omitempty" toml:"data" yaml:"data,omitempty"`
	CreatedAt    time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	ErrorData    null.JSON `boil:"error_data" json:"error_data,omitempty" toml:"error_data" yaml:"error_data,omitempty"`

	R *userDeviceDatumR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L userDeviceDatumL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserDeviceDatumColumns = struct {
	UserDeviceID string
	Data         string
	CreatedAt    string
	UpdatedAt    string
	ErrorData    string
}{
	UserDeviceID: "user_device_id",
	Data:         "data",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	ErrorData:    "error_data",
}

var UserDeviceDatumTableColumns = struct {
	UserDeviceID string
	Data         string
	CreatedAt    string
	UpdatedAt    string
	ErrorData    string
}{
	UserDeviceID: "user_device_data.user_device_id",
	Data:         "user_device_data.data",
	CreatedAt:    "user_device_data.created_at",
	UpdatedAt:    "user_device_data.updated_at",
	ErrorData:    "user_device_data.error_data",
}

// Generated where

var UserDeviceDatumWhere = struct {
	UserDeviceID whereHelperstring
	Data         whereHelpernull_JSON
	CreatedAt    whereHelpertime_Time
	UpdatedAt    whereHelpertime_Time
	ErrorData    whereHelpernull_JSON
}{
	UserDeviceID: whereHelperstring{field: "\"devices_api\".\"user_device_data\".\"user_device_id\""},
	Data:         whereHelpernull_JSON{field: "\"devices_api\".\"user_device_data\".\"data\""},
	CreatedAt:    whereHelpertime_Time{field: "\"devices_api\".\"user_device_data\".\"created_at\""},
	UpdatedAt:    whereHelpertime_Time{field: "\"devices_api\".\"user_device_data\".\"updated_at\""},
	ErrorData:    whereHelpernull_JSON{field: "\"devices_api\".\"user_device_data\".\"error_data\""},
}

// UserDeviceDatumRels is where relationship names are stored.
var UserDeviceDatumRels = struct {
	UserDevice string
}{
	UserDevice: "UserDevice",
}

// userDeviceDatumR is where relationships are stored.
type userDeviceDatumR struct {
	UserDevice *UserDevice `boil:"UserDevice" json:"UserDevice" toml:"UserDevice" yaml:"UserDevice"`
}

// NewStruct creates a new relationship struct
func (*userDeviceDatumR) NewStruct() *userDeviceDatumR {
	return &userDeviceDatumR{}
}

// userDeviceDatumL is where Load methods for each relationship are stored.
type userDeviceDatumL struct{}

var (
	userDeviceDatumAllColumns            = []string{"user_device_id", "data", "created_at", "updated_at", "error_data"}
	userDeviceDatumColumnsWithoutDefault = []string{"user_device_id"}
	userDeviceDatumColumnsWithDefault    = []string{"data", "created_at", "updated_at", "error_data"}
	userDeviceDatumPrimaryKeyColumns     = []string{"user_device_id"}
	userDeviceDatumGeneratedColumns      = []string{}
)

type (
	// UserDeviceDatumSlice is an alias for a slice of pointers to UserDeviceDatum.
	// This should almost always be used instead of []UserDeviceDatum.
	UserDeviceDatumSlice []*UserDeviceDatum
	// UserDeviceDatumHook is the signature for custom UserDeviceDatum hook methods
	UserDeviceDatumHook func(context.Context, boil.ContextExecutor, *UserDeviceDatum) error

	userDeviceDatumQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userDeviceDatumType                 = reflect.TypeOf(&UserDeviceDatum{})
	userDeviceDatumMapping              = queries.MakeStructMapping(userDeviceDatumType)
	userDeviceDatumPrimaryKeyMapping, _ = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, userDeviceDatumPrimaryKeyColumns)
	userDeviceDatumInsertCacheMut       sync.RWMutex
	userDeviceDatumInsertCache          = make(map[string]insertCache)
	userDeviceDatumUpdateCacheMut       sync.RWMutex
	userDeviceDatumUpdateCache          = make(map[string]updateCache)
	userDeviceDatumUpsertCacheMut       sync.RWMutex
	userDeviceDatumUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userDeviceDatumAfterSelectHooks []UserDeviceDatumHook

var userDeviceDatumBeforeInsertHooks []UserDeviceDatumHook
var userDeviceDatumAfterInsertHooks []UserDeviceDatumHook

var userDeviceDatumBeforeUpdateHooks []UserDeviceDatumHook
var userDeviceDatumAfterUpdateHooks []UserDeviceDatumHook

var userDeviceDatumBeforeDeleteHooks []UserDeviceDatumHook
var userDeviceDatumAfterDeleteHooks []UserDeviceDatumHook

var userDeviceDatumBeforeUpsertHooks []UserDeviceDatumHook
var userDeviceDatumAfterUpsertHooks []UserDeviceDatumHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserDeviceDatum) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserDeviceDatum) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserDeviceDatum) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserDeviceDatum) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserDeviceDatum) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserDeviceDatum) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserDeviceDatum) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserDeviceDatum) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserDeviceDatum) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range userDeviceDatumAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserDeviceDatumHook registers your hook function for all future operations.
func AddUserDeviceDatumHook(hookPoint boil.HookPoint, userDeviceDatumHook UserDeviceDatumHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		userDeviceDatumAfterSelectHooks = append(userDeviceDatumAfterSelectHooks, userDeviceDatumHook)
	case boil.BeforeInsertHook:
		userDeviceDatumBeforeInsertHooks = append(userDeviceDatumBeforeInsertHooks, userDeviceDatumHook)
	case boil.AfterInsertHook:
		userDeviceDatumAfterInsertHooks = append(userDeviceDatumAfterInsertHooks, userDeviceDatumHook)
	case boil.BeforeUpdateHook:
		userDeviceDatumBeforeUpdateHooks = append(userDeviceDatumBeforeUpdateHooks, userDeviceDatumHook)
	case boil.AfterUpdateHook:
		userDeviceDatumAfterUpdateHooks = append(userDeviceDatumAfterUpdateHooks, userDeviceDatumHook)
	case boil.BeforeDeleteHook:
		userDeviceDatumBeforeDeleteHooks = append(userDeviceDatumBeforeDeleteHooks, userDeviceDatumHook)
	case boil.AfterDeleteHook:
		userDeviceDatumAfterDeleteHooks = append(userDeviceDatumAfterDeleteHooks, userDeviceDatumHook)
	case boil.BeforeUpsertHook:
		userDeviceDatumBeforeUpsertHooks = append(userDeviceDatumBeforeUpsertHooks, userDeviceDatumHook)
	case boil.AfterUpsertHook:
		userDeviceDatumAfterUpsertHooks = append(userDeviceDatumAfterUpsertHooks, userDeviceDatumHook)
	}
}

// One returns a single userDeviceDatum record from the query.
func (q userDeviceDatumQuery) One(ctx context.Context, exec boil.ContextExecutor) (*UserDeviceDatum, error) {
	o := &UserDeviceDatum{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for user_device_data")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserDeviceDatum records from the query.
func (q userDeviceDatumQuery) All(ctx context.Context, exec boil.ContextExecutor) (UserDeviceDatumSlice, error) {
	var o []*UserDeviceDatum

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to UserDeviceDatum slice")
	}

	if len(userDeviceDatumAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserDeviceDatum records in the query.
func (q userDeviceDatumQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count user_device_data rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userDeviceDatumQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if user_device_data exists")
	}

	return count > 0, nil
}

// UserDevice pointed to by the foreign key.
func (o *UserDeviceDatum) UserDevice(mods ...qm.QueryMod) userDeviceQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserDeviceID),
	}

	queryMods = append(queryMods, mods...)

	query := UserDevices(queryMods...)
	queries.SetFrom(query.Query, "\"devices_api\".\"user_devices\"")

	return query
}

// LoadUserDevice allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (userDeviceDatumL) LoadUserDevice(ctx context.Context, e boil.ContextExecutor, singular bool, maybeUserDeviceDatum interface{}, mods queries.Applicator) error {
	var slice []*UserDeviceDatum
	var object *UserDeviceDatum

	if singular {
		object = maybeUserDeviceDatum.(*UserDeviceDatum)
	} else {
		slice = *maybeUserDeviceDatum.(*[]*UserDeviceDatum)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userDeviceDatumR{}
		}
		args = append(args, object.UserDeviceID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userDeviceDatumR{}
			}

			for _, a := range args {
				if a == obj.UserDeviceID {
					continue Outer
				}
			}

			args = append(args, obj.UserDeviceID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`devices_api.user_devices`),
		qm.WhereIn(`devices_api.user_devices.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load UserDevice")
	}

	var resultSlice []*UserDevice
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice UserDevice")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for user_devices")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user_devices")
	}

	if len(userDeviceDatumAfterSelectHooks) != 0 {
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
		object.R.UserDevice = foreign
		if foreign.R == nil {
			foreign.R = &userDeviceR{}
		}
		foreign.R.UserDeviceDatum = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserDeviceID == foreign.ID {
				local.R.UserDevice = foreign
				if foreign.R == nil {
					foreign.R = &userDeviceR{}
				}
				foreign.R.UserDeviceDatum = local
				break
			}
		}
	}

	return nil
}

// SetUserDevice of the userDeviceDatum to the related item.
// Sets o.R.UserDevice to related.
// Adds o to related.R.UserDeviceDatum.
func (o *UserDeviceDatum) SetUserDevice(ctx context.Context, exec boil.ContextExecutor, insert bool, related *UserDevice) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"user_device_data\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_device_id"}),
		strmangle.WhereClause("\"", "\"", 2, userDeviceDatumPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.UserDeviceID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UserDeviceID = related.ID
	if o.R == nil {
		o.R = &userDeviceDatumR{
			UserDevice: related,
		}
	} else {
		o.R.UserDevice = related
	}

	if related.R == nil {
		related.R = &userDeviceR{
			UserDeviceDatum: o,
		}
	} else {
		related.R.UserDeviceDatum = o
	}

	return nil
}

// UserDeviceData retrieves all the records using an executor.
func UserDeviceData(mods ...qm.QueryMod) userDeviceDatumQuery {
	mods = append(mods, qm.From("\"devices_api\".\"user_device_data\""))
	return userDeviceDatumQuery{NewQuery(mods...)}
}

// FindUserDeviceDatum retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserDeviceDatum(ctx context.Context, exec boil.ContextExecutor, userDeviceID string, selectCols ...string) (*UserDeviceDatum, error) {
	userDeviceDatumObj := &UserDeviceDatum{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"user_device_data\" where \"user_device_id\"=$1", sel,
	)

	q := queries.Raw(query, userDeviceID)

	err := q.Bind(ctx, exec, userDeviceDatumObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from user_device_data")
	}

	if err = userDeviceDatumObj.doAfterSelectHooks(ctx, exec); err != nil {
		return userDeviceDatumObj, err
	}

	return userDeviceDatumObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserDeviceDatum) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no user_device_data provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(userDeviceDatumColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userDeviceDatumInsertCacheMut.RLock()
	cache, cached := userDeviceDatumInsertCache[key]
	userDeviceDatumInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userDeviceDatumAllColumns,
			userDeviceDatumColumnsWithDefault,
			userDeviceDatumColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"user_device_data\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"user_device_data\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into user_device_data")
	}

	if !cached {
		userDeviceDatumInsertCacheMut.Lock()
		userDeviceDatumInsertCache[key] = cache
		userDeviceDatumInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the UserDeviceDatum.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserDeviceDatum) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userDeviceDatumUpdateCacheMut.RLock()
	cache, cached := userDeviceDatumUpdateCache[key]
	userDeviceDatumUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userDeviceDatumAllColumns,
			userDeviceDatumPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update user_device_data, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"user_device_data\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, userDeviceDatumPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, append(wl, userDeviceDatumPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update user_device_data row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for user_device_data")
	}

	if !cached {
		userDeviceDatumUpdateCacheMut.Lock()
		userDeviceDatumUpdateCache[key] = cache
		userDeviceDatumUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userDeviceDatumQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for user_device_data")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for user_device_data")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserDeviceDatumSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDeviceDatumPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"user_device_data\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, userDeviceDatumPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in userDeviceDatum slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all userDeviceDatum")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserDeviceDatum) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no user_device_data provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(userDeviceDatumColumnsWithDefault, o)

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

	userDeviceDatumUpsertCacheMut.RLock()
	cache, cached := userDeviceDatumUpsertCache[key]
	userDeviceDatumUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userDeviceDatumAllColumns,
			userDeviceDatumColumnsWithDefault,
			userDeviceDatumColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			userDeviceDatumAllColumns,
			userDeviceDatumPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert user_device_data, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(userDeviceDatumPrimaryKeyColumns))
			copy(conflict, userDeviceDatumPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"user_device_data\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userDeviceDatumType, userDeviceDatumMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert user_device_data")
	}

	if !cached {
		userDeviceDatumUpsertCacheMut.Lock()
		userDeviceDatumUpsertCache[key] = cache
		userDeviceDatumUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single UserDeviceDatum record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserDeviceDatum) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no UserDeviceDatum provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userDeviceDatumPrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"user_device_data\" WHERE \"user_device_id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from user_device_data")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for user_device_data")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userDeviceDatumQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no userDeviceDatumQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from user_device_data")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for user_device_data")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserDeviceDatumSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userDeviceDatumBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDeviceDatumPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"user_device_data\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userDeviceDatumPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from userDeviceDatum slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for user_device_data")
	}

	if len(userDeviceDatumAfterDeleteHooks) != 0 {
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
func (o *UserDeviceDatum) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindUserDeviceDatum(ctx, exec, o.UserDeviceID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserDeviceDatumSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserDeviceDatumSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userDeviceDatumPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"user_device_data\".* FROM \"devices_api\".\"user_device_data\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userDeviceDatumPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in UserDeviceDatumSlice")
	}

	*o = slice

	return nil
}

// UserDeviceDatumExists checks if the UserDeviceDatum row exists.
func UserDeviceDatumExists(ctx context.Context, exec boil.ContextExecutor, userDeviceID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"user_device_data\" where \"user_device_id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, userDeviceID)
	}
	row := exec.QueryRowContext(ctx, sql, userDeviceID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if user_device_data exists")
	}

	return exists, nil
}
