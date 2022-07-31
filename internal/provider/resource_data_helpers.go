package provider

import (
	"reflect"
	"strconv"

	"github.com/auth0/go-auth0"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

// ResourceData generalises schema.ResourceData so that we can reuse the
// accessor methods defined below.
type ResourceData interface {
	// IsNewResource reports whether or not the resource is seen for the first
	// time. If so, checks for change won't be carried out.
	IsNewResource() bool

	// HasChange reports whether or not the given key has been changed.
	HasChange(key string) bool

	// GetChange returns the old and new value for a given key.
	GetChange(key string) (interface{}, interface{})

	// Get returns the data for the given key, or nil if the key doesn't exist
	// in the schema.
	Get(key string) interface{}

	// GetOk returns the data for the given key and whether or not the key
	// has been set to a non-zero value at some point.
	//
	// The first result will not necessarilly be nil if the value doesn't exist.
	// The second result should be checked to determine this information.
	GetOk(key string) (interface{}, bool)

	// GetOkExists can check if TypeBool attributes that are Optional with
	// no Default value have been set.
	//
	// Deprecated: usage is discouraged due to undefined behaviors and may be
	// removed in a future version of the SDK
	GetOkExists(key string) (interface{}, bool)

	// Set sets the value for the given key.
	//
	// If the key is invalid or the value is not a correct type, an error
	// will be returned.
	Set(key string, value interface{}) error
}

type resourceData struct {
	ResourceData
	prefix string
}

func newResourceDataAtKey(key string, d ResourceData) ResourceData {
	return &resourceData{d, key}
}

func newResourceDataAtIndex(i int, d ResourceData) ResourceData {
	return &resourceData{d, strconv.Itoa(i)}
}

// IsNewResource reports whether the resource is seen for the first time.
func (d *resourceData) IsNewResource() bool {
	return d.ResourceData.IsNewResource()
}

// HasChange reports whether the given key has been changed.
func (d *resourceData) HasChange(key string) bool {
	return d.ResourceData.HasChange(d.prefix + "." + key)
}

// GetChange returns the old and new value for a given key.
func (d *resourceData) GetChange(key string) (interface{}, interface{}) {
	return d.ResourceData.GetChange(d.prefix + "." + key)
}

// Get returns the data for the given key, or nil
// if the key doesn't exist in the schema.
func (d *resourceData) Get(key string) interface{} {
	return d.ResourceData.Get(d.prefix + "." + key)
}

// GetOk returns the data for the given key and whether the
// key has been set to a non-zero value at some point.
func (d *resourceData) GetOk(key string) (interface{}, bool) {
	return d.ResourceData.GetOk(d.prefix + "." + key)
}

// GetOkExists can check if TypeBool attributes that are
// Optional with no Default value have been set.
func (d *resourceData) GetOkExists(key string) (interface{}, bool) {
	return d.ResourceData.GetOkExists(d.prefix + "." + key)
}

// Set sets the value for the given key.
func (d *resourceData) Set(key string, value interface{}) error {
	return d.ResourceData.Set(d.prefix+"."+key, value)
}

// MapData wraps a map satisfying the Data interface, so it can be used in the
// accessor methods defined below.
type MapData map[string]interface{}

// IsNewResource reports whether the resource is seen for the first time.
func (md MapData) IsNewResource() bool {
	return false
}

// HasChange reports whether the given key has been changed.
func (md MapData) HasChange(key string) bool {
	_, ok := md[key]
	return ok
}

// GetChange returns the old and new value for a given key.
func (md MapData) GetChange(key string) (interface{}, interface{}) {
	return md[key], md[key]
}

// Get returns the data for the given key, or nil
// if the key doesn't exist in the schema.
func (md MapData) Get(key string) interface{} {
	return md[key]
}

// GetOk returns the data for the given key and whether the
// key has been set to a non-zero value at some point.
func (md MapData) GetOk(key string) (interface{}, bool) {
	v, ok := md[key]
	return v, ok && !isNil(v) && !isZero(v)
}

// GetOkExists can check if TypeBool attributes that are
// Optional with no Default value have been set.
func (md MapData) GetOkExists(key string) (interface{}, bool) {
	v, ok := md[key]
	return v, ok && !isNil(v)
}

// Set sets the value for the given key.
func (md MapData) Set(key string, value interface{}) error {
	if !isNil(value) {
		md[key] = value
	}
	return nil
}

func isNil(v interface{}) bool {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map:
		return rv.IsNil()
	}
	return v == nil
}

func isZero(v interface{}) bool {
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

var _ ResourceData = (*schema.ResourceData)(nil)

// Condition is a function that checks whether a condition holds true for a
// value being accessed.
//
// It is used with accessor functions such as Int, String, etc to only retrieve
// the value if the conditions hold true.
type Condition func(d ResourceData, key string) bool

// Eval performs the evaluation of the condition.
func (c Condition) Eval(d ResourceData, key string) bool {
	return c(d, key)
}

// IsNewResource is a condition that evaluates to true if the resource access is
// new.
func IsNewResource() Condition {
	return func(d ResourceData, key string) bool {
		return d.IsNewResource()
	}
}

// HasChange is a condition that evaluates to true if the value accessed has
// changed.
func HasChange() Condition {
	return func(d ResourceData, key string) bool {
		return d.HasChange(key)
	}
}

// Any is a condition that evaluates to true if any of its enclosed conditions
// evaluate to true. If it is not passed any conditions it will be considered
// unconditional, therefore it will evaluate to true.
func Any(conditions ...Condition) Condition {
	return func(d ResourceData, key string) bool {
		for _, condition := range conditions {
			if condition.Eval(d, key) {
				return true
			}
		}
		return len(conditions) == 0
	}
}

// All is a condition that evaluates to true if all of its child conditions
// evaluate to true.
func All(conditions ...Condition) Condition {
	return func(d ResourceData, key string) bool {
		for _, condition := range conditions {
			if !condition.Eval(d, key) {
				return false
			}
		}
		return true
	}
}

// Not is a condition that evaluates to true if its child condition evaluates to
// false. False otherwise.
func Not(condition Condition) Condition {
	return func(d ResourceData, key string) bool {
		return !condition.Eval(d, key)
	}
}

// String accesses the value held by key and type asserts it to a pointer to a
// string.
func String(d ResourceData, key string, conditions ...Condition) (s *string) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		s = auth0.String(v.(string))
	}
	return
}

// Int accesses the value held by key and type asserts it to a pointer to a
// int.
func Int(d ResourceData, key string, conditions ...Condition) (i *int) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		i = auth0.Int(v.(int))
	}
	return
}

// Float64 accesses the value held by key and type asserts it to a pointer to a
// float64.
func Float64(d ResourceData, key string, conditions ...Condition) (f *float64) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		f = auth0.Float64(v.(float64))
	}
	return
}

// Bool accesses the value held by key and type asserts it to a pointer to a
// bool.
func Bool(d ResourceData, key string, conditions ...Condition) (b *bool) {
	// We need to keep using this for bool values even if it's marked as deprecated.
	// See https://github.com/hashicorp/terraform-plugin-sdk/issues/817.
	v, ok := d.GetOkExists(key)
	if ok && Any(conditions...).Eval(d, key) {
		b = auth0.Bool(v.(bool))
	}
	return
}

// Flag accesses the value held by the bool flag key
// and type asserts it as a pointer to a bool.
func Flag(configFlags cty.Value, key string) *bool {
	configFlag := configFlags.GetAttr(key)

	if configFlag.IsNull() {
		return nil
	}

	return auth0.Bool(configFlag.True())
}

// Slice accesses the value held by key and type asserts it to a slice.
func Slice(d ResourceData, key string, conditions ...Condition) (s []interface{}) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		s = v.([]interface{})
	}
	return
}

// Map accesses the value held by key and type asserts it to a map.
func Map(d ResourceData, key string, conditions ...Condition) (m map[string]interface{}) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		m = v.(map[string]interface{})
	}
	return
}

// List accesses the value held by key and returns an iterator able to go over
// its elements.
func List(d ResourceData, key string, conditions ...Condition) Iterator {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		return &list{newResourceDataAtKey(key, d), v.([]interface{})}
	}
	return &list{}
}

// Set accesses the value held by key, type asserts it to a set and returns an
// iterator able to go over its elements.
func Set(d ResourceData, key string, conditions ...Condition) Iterator {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		if s, ok := v.(*schema.Set); ok {
			return &set{newResourceDataAtKey(key, d), s}
		}
	}
	return &set{nil, &schema.Set{}}
}

// Iterator is used to iterate over a list or set.
//
// Elem iterates over all elements of the iterator, calling fn with each
// iteration. The callback takes a Data interface as argument which is prefixed
// with its parents key, allowing for convenient nested data access.
//
// List returns all elements of the iterator as a Go slice.
//
// Len returns the number of elements held by the iterator.
type Iterator interface {
	Elem(func(d ResourceData))
	List() []interface{}
	Len() int
}

type list struct {
	d ResourceData
	v []interface{}
}

// Elem will loop through each element of the list and apply fn().
func (l *list) Elem(fn func(ResourceData)) {
	for idx := range l.v {
		fn(newResourceDataAtIndex(idx, l.d))
	}
}

// List will return the values as list.
func (l *list) List() []interface{} {
	return l.v
}

// Len returns the length of the list values.
func (l *list) Len() int {
	return len(l.v)
}

type set struct {
	d ResourceData
	s *schema.Set
}

func (s *set) hash(item interface{}) string {
	code := s.s.F(item)
	if code < 0 {
		code = -code
	}
	return strconv.Itoa(code)
}

// Elem will loop through each element of the set and apply fn().
func (s *set) Elem(fn func(ResourceData)) {
	for _, v := range s.s.List() {
		fn(newResourceDataAtKey(s.hash(v), s.d))
	}
}

// List will return the values of the set as a list.
func (s *set) List() []interface{} {
	return s.s.List()
}

// Len returns the length of the set values.
func (s *set) Len() int {
	return s.s.Len()
}

// Diff accesses the value held by key and type asserts it to a set. It then
// compares its changes if any and returns what needs to be added and what
// needs to be removed.
func Diff(d ResourceData, key string) (add Iterator, rm Iterator) {
	// Zero the add and rm sets. These may be modified if the diff observed any
	// changes.
	add = &set{newResourceDataAtKey(key, d), d.Get(key).(*schema.Set)}
	rm = &set{newResourceDataAtKey(key, d), &schema.Set{}}

	if d.HasChange(key) {
		o, n := d.GetChange(key)
		add = &set{newResourceDataAtKey(key, d), n.(*schema.Set).Difference(o.(*schema.Set))}
		rm = &set{newResourceDataAtKey(key, d), o.(*schema.Set).Difference(n.(*schema.Set))}
	}
	return
}

// JSON accesses the value held by key and unmarshals it into a map.
func JSON(d ResourceData, key string, conditions ...Condition) (m map[string]interface{}, err error) {
	v, ok := d.GetOk(key)
	if ok && Any(conditions...).Eval(d, key) {
		m, err = structure.ExpandJsonFromString(v.(string))
		if err != nil {
			return
		}
	}
	return
}
