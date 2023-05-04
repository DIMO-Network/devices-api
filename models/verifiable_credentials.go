// Code generated by SQLBoiler 4.14.1 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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

// VerifiableCredential is an object representing the database table.
type VerifiableCredential struct {
	UserDeviceID   string            `boil:"user_device_id" json:"user_device_id" toml:"user_device_id" yaml:"user_device_id"`
	Vin            string            `boil:"vin" json:"vin" toml:"vin" yaml:"vin"`
	TokenID        types.NullDecimal `boil:"token_id" json:"token_id,omitempty" toml:"token_id" yaml:"token_id,omitempty"`
	ClaimsRoot     null.Bytes        `boil:"claims_root" json:"claims_root,omitempty" toml:"claims_root" yaml:"claims_root,omitempty"`
	RevocationRoot null.Bytes        `boil:"revocation_root" json:"revocation_root,omitempty" toml:"revocation_root" yaml:"revocation_root,omitempty"`
	RootOfRoots    null.Bytes        `boil:"root_of_roots" json:"root_of_roots,omitempty" toml:"root_of_roots" yaml:"root_of_roots,omitempty"`
	State          types.NullDecimal `boil:"state" json:"state,omitempty" toml:"state" yaml:"state,omitempty"`

	R *verifiableCredentialR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L verifiableCredentialL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var VerifiableCredentialColumns = struct {
	UserDeviceID   string
	Vin            string
	TokenID        string
	ClaimsRoot     string
	RevocationRoot string
	RootOfRoots    string
	State          string
}{
	UserDeviceID:   "user_device_id",
	Vin:            "vin",
	TokenID:        "token_id",
	ClaimsRoot:     "claims_root",
	RevocationRoot: "revocation_root",
	RootOfRoots:    "root_of_roots",
	State:          "state",
}

var VerifiableCredentialTableColumns = struct {
	UserDeviceID   string
	Vin            string
	TokenID        string
	ClaimsRoot     string
	RevocationRoot string
	RootOfRoots    string
	State          string
}{
	UserDeviceID:   "verifiable_credentials.user_device_id",
	Vin:            "verifiable_credentials.vin",
	TokenID:        "verifiable_credentials.token_id",
	ClaimsRoot:     "verifiable_credentials.claims_root",
	RevocationRoot: "verifiable_credentials.revocation_root",
	RootOfRoots:    "verifiable_credentials.root_of_roots",
	State:          "verifiable_credentials.state",
}

// Generated where

var VerifiableCredentialWhere = struct {
	UserDeviceID   whereHelperstring
	Vin            whereHelperstring
	TokenID        whereHelpertypes_NullDecimal
	ClaimsRoot     whereHelpernull_Bytes
	RevocationRoot whereHelpernull_Bytes
	RootOfRoots    whereHelpernull_Bytes
	State          whereHelpertypes_NullDecimal
}{
	UserDeviceID:   whereHelperstring{field: "\"devices_api\".\"verifiable_credentials\".\"user_device_id\""},
	Vin:            whereHelperstring{field: "\"devices_api\".\"verifiable_credentials\".\"vin\""},
	TokenID:        whereHelpertypes_NullDecimal{field: "\"devices_api\".\"verifiable_credentials\".\"token_id\""},
	ClaimsRoot:     whereHelpernull_Bytes{field: "\"devices_api\".\"verifiable_credentials\".\"claims_root\""},
	RevocationRoot: whereHelpernull_Bytes{field: "\"devices_api\".\"verifiable_credentials\".\"revocation_root\""},
	RootOfRoots:    whereHelpernull_Bytes{field: "\"devices_api\".\"verifiable_credentials\".\"root_of_roots\""},
	State:          whereHelpertypes_NullDecimal{field: "\"devices_api\".\"verifiable_credentials\".\"state\""},
}

// VerifiableCredentialRels is where relationship names are stored.
var VerifiableCredentialRels = struct {
}{}

// verifiableCredentialR is where relationships are stored.
type verifiableCredentialR struct {
}

// NewStruct creates a new relationship struct
func (*verifiableCredentialR) NewStruct() *verifiableCredentialR {
	return &verifiableCredentialR{}
}

// verifiableCredentialL is where Load methods for each relationship are stored.
type verifiableCredentialL struct{}

var (
	verifiableCredentialAllColumns            = []string{"user_device_id", "vin", "token_id", "claims_root", "revocation_root", "root_of_roots", "state"}
	verifiableCredentialColumnsWithoutDefault = []string{"user_device_id", "vin"}
	verifiableCredentialColumnsWithDefault    = []string{"token_id", "claims_root", "revocation_root", "root_of_roots", "state"}
	verifiableCredentialPrimaryKeyColumns     = []string{"user_device_id"}
	verifiableCredentialGeneratedColumns      = []string{}
)

type (
	// VerifiableCredentialSlice is an alias for a slice of pointers to VerifiableCredential.
	// This should almost always be used instead of []VerifiableCredential.
	VerifiableCredentialSlice []*VerifiableCredential
	// VerifiableCredentialHook is the signature for custom VerifiableCredential hook methods
	VerifiableCredentialHook func(context.Context, boil.ContextExecutor, *VerifiableCredential) error

	verifiableCredentialQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	verifiableCredentialType                 = reflect.TypeOf(&VerifiableCredential{})
	verifiableCredentialMapping              = queries.MakeStructMapping(verifiableCredentialType)
	verifiableCredentialPrimaryKeyMapping, _ = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, verifiableCredentialPrimaryKeyColumns)
	verifiableCredentialInsertCacheMut       sync.RWMutex
	verifiableCredentialInsertCache          = make(map[string]insertCache)
	verifiableCredentialUpdateCacheMut       sync.RWMutex
	verifiableCredentialUpdateCache          = make(map[string]updateCache)
	verifiableCredentialUpsertCacheMut       sync.RWMutex
	verifiableCredentialUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var verifiableCredentialAfterSelectHooks []VerifiableCredentialHook

var verifiableCredentialBeforeInsertHooks []VerifiableCredentialHook
var verifiableCredentialAfterInsertHooks []VerifiableCredentialHook

var verifiableCredentialBeforeUpdateHooks []VerifiableCredentialHook
var verifiableCredentialAfterUpdateHooks []VerifiableCredentialHook

var verifiableCredentialBeforeDeleteHooks []VerifiableCredentialHook
var verifiableCredentialAfterDeleteHooks []VerifiableCredentialHook

var verifiableCredentialBeforeUpsertHooks []VerifiableCredentialHook
var verifiableCredentialAfterUpsertHooks []VerifiableCredentialHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *VerifiableCredential) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *VerifiableCredential) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *VerifiableCredential) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *VerifiableCredential) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *VerifiableCredential) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *VerifiableCredential) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *VerifiableCredential) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *VerifiableCredential) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *VerifiableCredential) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range verifiableCredentialAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddVerifiableCredentialHook registers your hook function for all future operations.
func AddVerifiableCredentialHook(hookPoint boil.HookPoint, verifiableCredentialHook VerifiableCredentialHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		verifiableCredentialAfterSelectHooks = append(verifiableCredentialAfterSelectHooks, verifiableCredentialHook)
	case boil.BeforeInsertHook:
		verifiableCredentialBeforeInsertHooks = append(verifiableCredentialBeforeInsertHooks, verifiableCredentialHook)
	case boil.AfterInsertHook:
		verifiableCredentialAfterInsertHooks = append(verifiableCredentialAfterInsertHooks, verifiableCredentialHook)
	case boil.BeforeUpdateHook:
		verifiableCredentialBeforeUpdateHooks = append(verifiableCredentialBeforeUpdateHooks, verifiableCredentialHook)
	case boil.AfterUpdateHook:
		verifiableCredentialAfterUpdateHooks = append(verifiableCredentialAfterUpdateHooks, verifiableCredentialHook)
	case boil.BeforeDeleteHook:
		verifiableCredentialBeforeDeleteHooks = append(verifiableCredentialBeforeDeleteHooks, verifiableCredentialHook)
	case boil.AfterDeleteHook:
		verifiableCredentialAfterDeleteHooks = append(verifiableCredentialAfterDeleteHooks, verifiableCredentialHook)
	case boil.BeforeUpsertHook:
		verifiableCredentialBeforeUpsertHooks = append(verifiableCredentialBeforeUpsertHooks, verifiableCredentialHook)
	case boil.AfterUpsertHook:
		verifiableCredentialAfterUpsertHooks = append(verifiableCredentialAfterUpsertHooks, verifiableCredentialHook)
	}
}

// One returns a single verifiableCredential record from the query.
func (q verifiableCredentialQuery) One(ctx context.Context, exec boil.ContextExecutor) (*VerifiableCredential, error) {
	o := &VerifiableCredential{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for verifiable_credentials")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all VerifiableCredential records from the query.
func (q verifiableCredentialQuery) All(ctx context.Context, exec boil.ContextExecutor) (VerifiableCredentialSlice, error) {
	var o []*VerifiableCredential

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to VerifiableCredential slice")
	}

	if len(verifiableCredentialAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all VerifiableCredential records in the query.
func (q verifiableCredentialQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count verifiable_credentials rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q verifiableCredentialQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if verifiable_credentials exists")
	}

	return count > 0, nil
}

// VerifiableCredentials retrieves all the records using an executor.
func VerifiableCredentials(mods ...qm.QueryMod) verifiableCredentialQuery {
	mods = append(mods, qm.From("\"devices_api\".\"verifiable_credentials\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"devices_api\".\"verifiable_credentials\".*"})
	}

	return verifiableCredentialQuery{q}
}

// FindVerifiableCredential retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindVerifiableCredential(ctx context.Context, exec boil.ContextExecutor, userDeviceID string, selectCols ...string) (*VerifiableCredential, error) {
	verifiableCredentialObj := &VerifiableCredential{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices_api\".\"verifiable_credentials\" where \"user_device_id\"=$1", sel,
	)

	q := queries.Raw(query, userDeviceID)

	err := q.Bind(ctx, exec, verifiableCredentialObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from verifiable_credentials")
	}

	if err = verifiableCredentialObj.doAfterSelectHooks(ctx, exec); err != nil {
		return verifiableCredentialObj, err
	}

	return verifiableCredentialObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *VerifiableCredential) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no verifiable_credentials provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(verifiableCredentialColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	verifiableCredentialInsertCacheMut.RLock()
	cache, cached := verifiableCredentialInsertCache[key]
	verifiableCredentialInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			verifiableCredentialAllColumns,
			verifiableCredentialColumnsWithDefault,
			verifiableCredentialColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices_api\".\"verifiable_credentials\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices_api\".\"verifiable_credentials\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "models: unable to insert into verifiable_credentials")
	}

	if !cached {
		verifiableCredentialInsertCacheMut.Lock()
		verifiableCredentialInsertCache[key] = cache
		verifiableCredentialInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the VerifiableCredential.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *VerifiableCredential) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	verifiableCredentialUpdateCacheMut.RLock()
	cache, cached := verifiableCredentialUpdateCache[key]
	verifiableCredentialUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			verifiableCredentialAllColumns,
			verifiableCredentialPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update verifiable_credentials, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices_api\".\"verifiable_credentials\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, verifiableCredentialPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, append(wl, verifiableCredentialPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update verifiable_credentials row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for verifiable_credentials")
	}

	if !cached {
		verifiableCredentialUpdateCacheMut.Lock()
		verifiableCredentialUpdateCache[key] = cache
		verifiableCredentialUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q verifiableCredentialQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for verifiable_credentials")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for verifiable_credentials")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o VerifiableCredentialSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), verifiableCredentialPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices_api\".\"verifiable_credentials\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, verifiableCredentialPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in verifiableCredential slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all verifiableCredential")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *VerifiableCredential) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no verifiable_credentials provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(verifiableCredentialColumnsWithDefault, o)

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

	verifiableCredentialUpsertCacheMut.RLock()
	cache, cached := verifiableCredentialUpsertCache[key]
	verifiableCredentialUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			verifiableCredentialAllColumns,
			verifiableCredentialColumnsWithDefault,
			verifiableCredentialColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			verifiableCredentialAllColumns,
			verifiableCredentialPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert verifiable_credentials, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(verifiableCredentialPrimaryKeyColumns))
			copy(conflict, verifiableCredentialPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices_api\".\"verifiable_credentials\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(verifiableCredentialType, verifiableCredentialMapping, ret)
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
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert verifiable_credentials")
	}

	if !cached {
		verifiableCredentialUpsertCacheMut.Lock()
		verifiableCredentialUpsertCache[key] = cache
		verifiableCredentialUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single VerifiableCredential record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *VerifiableCredential) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no VerifiableCredential provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), verifiableCredentialPrimaryKeyMapping)
	sql := "DELETE FROM \"devices_api\".\"verifiable_credentials\" WHERE \"user_device_id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from verifiable_credentials")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for verifiable_credentials")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q verifiableCredentialQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no verifiableCredentialQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from verifiable_credentials")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for verifiable_credentials")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o VerifiableCredentialSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(verifiableCredentialBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), verifiableCredentialPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"devices_api\".\"verifiable_credentials\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, verifiableCredentialPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from verifiableCredential slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for verifiable_credentials")
	}

	if len(verifiableCredentialAfterDeleteHooks) != 0 {
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
func (o *VerifiableCredential) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindVerifiableCredential(ctx, exec, o.UserDeviceID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *VerifiableCredentialSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := VerifiableCredentialSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), verifiableCredentialPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices_api\".\"verifiable_credentials\".* FROM \"devices_api\".\"verifiable_credentials\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, verifiableCredentialPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in VerifiableCredentialSlice")
	}

	*o = slice

	return nil
}

// VerifiableCredentialExists checks if the VerifiableCredential row exists.
func VerifiableCredentialExists(ctx context.Context, exec boil.ContextExecutor, userDeviceID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices_api\".\"verifiable_credentials\" where \"user_device_id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, userDeviceID)
	}
	row := exec.QueryRowContext(ctx, sql, userDeviceID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if verifiable_credentials exists")
	}

	return exists, nil
}

// Exists checks if the VerifiableCredential row exists.
func (o *VerifiableCredential) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	return VerifiableCredentialExists(ctx, exec, o.UserDeviceID)
}
