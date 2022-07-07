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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// MintRequest is an object representing the database table.
type MintRequest struct {
	ID           string            `boil:"id" json:"id" toml:"id" yaml:"id"`
	UserDeviceID string            `boil:"user_device_id" json:"user_device_id" toml:"user_device_id" yaml:"user_device_id"`
	TXState      string            `boil:"tx_state" json:"tx_state" toml:"tx_state" yaml:"tx_state"`
	TokenID      types.NullDecimal `boil:"token_id" json:"token_id,omitempty" toml:"token_id" yaml:"token_id,omitempty"`
	CreatedAt    time.Time         `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time         `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	TXHash       null.Bytes        `boil:"tx_hash" json:"tx_hash,omitempty" toml:"tx_hash" yaml:"tx_hash,omitempty"`

	R *mintRequestR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L mintRequestL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var MintRequestColumns = struct {
	ID           string
	UserDeviceID string
	TXState      string
	TokenID      string
	CreatedAt    string
	UpdatedAt    string
	TXHash       string
}{
	ID:           "id",
	UserDeviceID: "user_device_id",
	TXState:      "tx_state",
	TokenID:      "token_id",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	TXHash:       "tx_hash",
}

var MintRequestTableColumns = struct {
	ID           string
	UserDeviceID string
	TXState      string
	TokenID      string
	CreatedAt    string
	UpdatedAt    string
	TXHash       string
}{
	ID:           "mint_requests.id",
	UserDeviceID: "mint_requests.user_device_id",
	TXState:      "mint_requests.tx_state",
	TokenID:      "mint_requests.token_id",
	CreatedAt:    "mint_requests.created_at",
	UpdatedAt:    "mint_requests.updated_at",
	TXHash:       "mint_requests.tx_hash",
}

// Generated where

type whereHelpernull_Bytes struct{ field string }

func (w whereHelpernull_Bytes) EQ(x null.Bytes) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Bytes) NEQ(x null.Bytes) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Bytes) LT(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Bytes) LTE(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Bytes) GT(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Bytes) GTE(x null.Bytes) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Bytes) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Bytes) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var MintRequestWhere = struct {
	ID           whereHelperstring
	UserDeviceID whereHelperstring
	TXState      whereHelperstring
	TokenID      whereHelpertypes_NullDecimal
	CreatedAt    whereHelpertime_Time
	UpdatedAt    whereHelpertime_Time
	TXHash       whereHelpernull_Bytes
}{
	ID:           whereHelperstring{field: "\"devices_api\".\"mint_requests\".\"id\""},
	UserDeviceID: whereHelperstring{field: "\"devices_api\".\"mint_requests\".\"user_device_id\""},
	TXState:      whereHelperstring{field: "\"devices_api\".\"mint_requests\".\"tx_state\""},
	TokenID:      whereHelpertypes_NullDecimal{field: "\"devices_api\".\"mint_requests\".\"token_id\""},
	CreatedAt:    whereHelpertime_Time{field: "\"devices_api\".\"mint_requests\".\"created_at\""},
	UpdatedAt:    whereHelpertime_Time{field: "\"devices_api\".\"mint_requests\".\"updated_at\""},
	TXHash:       whereHelpernull_Bytes{field: "\"devices_api\".\"mint_requests\".\"tx_hash\""},
}

// MintRequestRels is where relationship names are stored.
var MintRequestRels = struct {
	UserDevice string
}{
	UserDevice: "UserDevice",
}

// mintRequestR is where relationships are stored.
type mintRequestR struct {
	UserDevice *UserDevice `boil:"UserDevice" json:"UserDevice" toml:"UserDevice" yaml:"UserDevice"`
}

// NewStruct creates a new relationship struct
func (*mintRequestR) NewStruct() *mintRequestR {
	return &mintRequestR{}
}

// mintRequestL is where Load methods for each relationship are stored.
type mintRequestL struct{}

var (
	mintRequestAllColumns            = []string{"id", "user_device_id", "tx_state", "token_id", "created_at", "updated_at", "tx_hash"}
	mintRequestColumnsWithoutDefault = []string{"id", "user_device_id"}
	mintRequestColumnsWithDefault    = []string{"tx_state", "token_id", "created_at", "updated_at", "tx_hash"}
	mintRequestPrimaryKeyColumns     = []string{"id"}
	mintRequestGeneratedColumns      = []string{}
)

type (
	// MintRequestSlice is an alias for a slice of pointers to MintRequest.
	// This should almost always be used instead of []MintRequest.
	MintRequestSlice []*MintRequest
	// MintRequestHook is the signature for custom MintRequest hook methods
	MintRequestHook func(context.Context, boil.ContextExecutor, *MintRequest) error

	mintRequestQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	mintRequestType                 = reflect.TypeOf(&MintRequest{})
	mintRequestMapping              = queries.MakeStructMapping(mintRequestType)
	mintRequestPrimaryKeyMapping, _ = queries.BindMapping(mintRequestType, mintRequestMapping, mintRequestPrimaryKeyColumns)
	mintRequestInsertCacheMut       sync.RWMutex
	mintRequestInsertCache          = make(map[string]insertCache)
	mintRequestUpdateCacheMut       sync.RWMutex
	mintRequestUpdateCache          = make(map[string]updateCache)
	mintRequestUpsertCacheMut       sync.RWMutex
	mintRequestUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var mintRequestAfterSelectHooks []MintRequestHook

var mintRequestBeforeInsertHooks []MintRequestHook
var mintRequestAfterInsertHooks []MintRequestHook

var mintRequestBeforeUpdateHooks []MintRequestHook
var mintRequestAfterUpdateHooks []MintRequestHook

var mintRequestBeforeDeleteHooks []MintRequestHook
var mintRequestAfterDeleteHooks []MintRequestHook

var mintRequestBeforeUpsertHooks []MintRequestHook
var mintRequestAfterUpsertHooks []MintRequestHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *MintRequest) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *MintRequest) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *MintRequest) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *MintRequest) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *MintRequest) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *MintRequest) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *MintRequest) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *MintRequest) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *MintRequest) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range mintRequestAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddMintRequestHook registers your hook function for all future operations.
func AddMintRequestHook(hookPoint boil.HookPoint, mintRequestHook MintRequestHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		mintRequestAfterSelectHooks = append(mintRequestAfterSelectHooks, mintRequestHook)
	case boil.BeforeInsertHook:
		mintRequestBeforeInsertHooks = append(mintRequestBeforeInsertHooks, mintRequestHook)
	case boil.AfterInsertHook:
		mintRequestAfterInsertHooks = append(mintRequestAfterInsertHooks, mintRequestHook)
	case boil.BeforeUpdateHook:
		mintRequestBeforeUpdateHooks = append(mintRequestBeforeUpdateHooks, mintRequestHook)
	case boil.AfterUpdateHook:
		mintRequestAfterUpdateHooks = append(mintRequestAfterUpdateHooks, mintRequestHook)
	case boil.BeforeDeleteHook:
		mintRequestBeforeDeleteHooks = append(mintRequestBeforeDeleteHooks, mintRequestHook)
	case boil.AfterDeleteHook:
		mintRequestAfterDeleteHooks = append(mintRequestAfterDeleteHooks, mintRequestHook)
	case boil.BeforeUpsertHook:
		mintRequestBeforeUpsertHooks = append(mintRequestBeforeUpsertHooks, mintRequestHook)
	case boil.AfterUpsertHook:
		mintRequestAfterUpsertHooks = append(mintRequestAfterUpsertHooks, mintRequestHook)
	}
}

// One returns a single mintRequest record from the query.
func (q mintRequestQuery) One(ctx context.Context, exec boil.ContextExecutor) (*MintRequest, error) {
	o := &MintRequest{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for mint_requests")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all MintRequest records from the query.
func (q mintRequestQuery) All(ctx context.Context, exec boil.ContextExecutor) (MintRequestSlice, error) {
	var o []*MintRequest

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to MintRequest slice")
	}

	if len(mintRequestAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all MintRequest records in the query.
func (q mintRequestQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count mint_requests rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q mintRequestQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if mint_requests exists")
	}

	return count > 0, nil
}

// UserDevice pointed to by the foreign key.
func (o *MintRequest) UserDevice(mods ...qm.QueryMod) userDeviceQuery {
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
func (mintRequestL) LoadUserDevice(ctx context.Context, e boil.ContextExecutor, singular bool, maybeMintRequest interface{}, mods queries.Applicator) error {
	var slice []*MintRequest
	var object *MintRequest

	if singular {
		object = maybeMintRequest.(*MintRequest)
	} else {
		slice = *maybeMintRequest.(*[]*MintRequest)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &mintRequestR{}
		}
		args = append(args, object.UserDeviceID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &mintRequestR{}
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

	if len(mintRequestAfterSelectHooks) != 0 {
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
		foreign.R.MintRequest = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserDeviceID == foreign.ID {
				local.R.UserDevice = foreign
				if foreign.R == nil {
					foreign.R = &userDeviceR{}
				}
				foreign.R.MintRequest = local
				break
			}
		}
	}

	return nil
}

// SetUserDevice of the mintRequest to the related item.
// Sets o.R.UserDevice to related.
// Adds o to related.R.MintRequest.
func (o *MintRequest) SetUserDevice(ctx context.Context, exec boil.ContextExecutor, insert bool, related *UserDevice) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices_api\".\"mint_requests\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_device_id"}),
		strmangle.WhereClause("\"", "\"", 2, mintRequestPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

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
		o.R = &mintRequestR{
			UserDevice: related,
		}
	} else {
		o.R.UserDevice = related
	}

	if related.R == nil {
		related.R = &userDeviceR{
			MintRequest: o,
		}
	} else {
		related.R.MintRequest = o
	}

	return nil
}

// MintRequests retrieves all the records using an executor.
func MintRequests(mods ...qm.QueryMod) mintRequestQuery {
	mods = append(mods, qm.From("\"devices_api\".\"mint_requests\""))
	return mintRequestQuery{NewQuery(mods...)}
}

// FindMintRequest retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindMintRequest(ctx context.Context, exec boil.ContextExecutor, iD string, selectCols ...string) (*MintRequest, error) {
	mintRequestObj := &MintRequest{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"mint_requests\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, mintRequestObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from mint_requests")
	}

	if err = mintRequestObj.doAfterSelectHooks(ctx, exec); err != nil {
		return mintRequestObj, err
	}

	return mintRequestObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *MintRequest) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no mint_requests provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(mintRequestColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	mintRequestInsertCacheMut.RLock()
	cache, cached := mintRequestInsertCache[key]
	mintRequestInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			mintRequestAllColumns,
			mintRequestColumnsWithDefault,
			mintRequestColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(mintRequestType, mintRequestMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(mintRequestType, mintRequestMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"mint_requests\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"mint_requests\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into mint_requests")
	}

	if !cached {
		mintRequestInsertCacheMut.Lock()
		mintRequestInsertCache[key] = cache
		mintRequestInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the MintRequest.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *MintRequest) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		o.UpdatedAt = currTime
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	mintRequestUpdateCacheMut.RLock()
	cache, cached := mintRequestUpdateCache[key]
	mintRequestUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			mintRequestAllColumns,
			mintRequestPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update mint_requests, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"mint_requests\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, mintRequestPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(mintRequestType, mintRequestMapping, append(wl, mintRequestPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update mint_requests row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for mint_requests")
	}

	if !cached {
		mintRequestUpdateCacheMut.Lock()
		mintRequestUpdateCache[key] = cache
		mintRequestUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q mintRequestQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for mint_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for mint_requests")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o MintRequestSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mintRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"mint_requests\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, mintRequestPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in mintRequest slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all mintRequest")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *MintRequest) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no mint_requests provided for upsert")
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

	nzDefaults := queries.NonZeroDefaultSet(mintRequestColumnsWithDefault, o)

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

	mintRequestUpsertCacheMut.RLock()
	cache, cached := mintRequestUpsertCache[key]
	mintRequestUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			mintRequestAllColumns,
			mintRequestColumnsWithDefault,
			mintRequestColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			mintRequestAllColumns,
			mintRequestPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert mint_requests, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(mintRequestPrimaryKeyColumns))
			copy(conflict, mintRequestPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"mint_requests\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(mintRequestType, mintRequestMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(mintRequestType, mintRequestMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert mint_requests")
	}

	if !cached {
		mintRequestUpsertCacheMut.Lock()
		mintRequestUpsertCache[key] = cache
		mintRequestUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single MintRequest record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *MintRequest) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no MintRequest provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), mintRequestPrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"mint_requests\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from mint_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for mint_requests")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q mintRequestQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no mintRequestQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from mint_requests")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for mint_requests")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o MintRequestSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(mintRequestBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mintRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"mint_requests\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, mintRequestPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from mintRequest slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for mint_requests")
	}

	if len(mintRequestAfterDeleteHooks) != 0 {
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
func (o *MintRequest) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindMintRequest(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *MintRequestSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := MintRequestSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mintRequestPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"mint_requests\".* FROM \"devices_api\".\"mint_requests\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, mintRequestPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in MintRequestSlice")
	}

	*o = slice

	return nil
}

// MintRequestExists checks if the MintRequest row exists.
func MintRequestExists(ctx context.Context, exec boil.ContextExecutor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"mint_requests\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if mint_requests exists")
	}

	return exists, nil
}
