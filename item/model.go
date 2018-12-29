package item

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/go-errors/errors"
)

// Model stores the known data model
type Model struct {
	sync.RWMutex
	TypeAttributes map[string]map[string]string
	typeChildren   map[string]map[string]struct{}
}

type modelOperation interface {
	apply(model *Model)
}

type addAttribute struct {
	itype string
	name  string
	atype string
}

func (add addAttribute) apply(model *Model) {
	m1 := model.TypeAttributes[add.itype]
	if m1 == nil {
		m1 = make(map[string]string)
	}
	m1[add.name] = add.atype
	model.TypeAttributes[add.itype] = m1
}

type addChild struct {
	ptype string
	ctype string
}

func (add addChild) apply(model *Model) {
	m1 := model.typeChildren[add.ptype]
	if m1 == nil {
		m1 = make(map[string]struct{})
	}
	m1[add.ctype] = struct{}{}
	model.typeChildren[add.ptype] = m1
}

// ModelID is the ID in the main store of the model
var ModelID = "Model"

// EmptyModel creates a new model
func EmptyModel() *Model {
	return &Model{TypeAttributes: make(map[string]map[string]string), typeChildren: make(map[string]map[string]struct{})}
}

// ToItem transforms the model in an item to save it in the store
func ToItem(model *Model) Item {
	cnts := make(map[string]interface{})
	model.RLock()
	childtypes := make(map[string][]string)
	for k := range model.typeChildren {
		childtypes[k] = childTypes(model, k)
	}
	cnts["typeChildren"] = childtypes

	attrs := make(map[string]map[string]string)
	for t, v := range model.TypeAttributes {
		tattr := make(map[string]string)
		for n, vt := range v {
			tattr[n] = vt
		}
		attrs[t] = tattr
	}
	cnts["typeAttributes"] = attrs
	model.RUnlock()
	return Item{ModelID, "Model", "Model", cnts}
}

// FromItem reads a model from an Item
func FromItem(item Item) *Model {
	model := EmptyModel()
	var ops []modelOperation
	cnts := item.Contents
	tc, ok := cnts["typeChildren"].(map[string][]string)
	if ok {
		for k, v := range tc {
			for _, t := range v {
				ops = append(ops, addChild{k, t})
			}
		}
	}
	ta, ok := cnts["typeAttributes"].(map[string]map[string]string)
	if ok {
		for k, v := range ta {
			for a, vt := range v {
				ops = append(ops, addAttribute{k, a, vt})
			}
		}
	}

	for _, op := range ops {
		op.apply(model)
	}
	return model
}

// AddItem registers the item model
func AddItem(item Item, model *Model) (bool, error) {
	err := checkID(item)
	if err != nil {
		return false, err
	}
	var errs []error
	var ops []modelOperation
	var changed bool
	model.RLock()
	m1 := model.TypeAttributes[item.Type]
	if m1 == nil {
		m1 = make(map[string]string)
	}
	for k, v := range item.Contents {
		vt := reflect.TypeOf(v).String()
		oldt, ok := m1[k]
		if !ok {
			ops = append(ops, addAttribute{item.Type, k, vt})
		} else if vt != oldt {
			errs = append(errs, errors.New(ModelError{"TYPE_MISMATCH",
				fmt.Sprintf("Attribute %s was %s, now %s", k, oldt, vt)}))
		}

	}

	ops = parentType(model, item.ID, item.Type, ops)
	model.RUnlock()
	if len(ops) > 0 {
		model.Lock()
		for _, op := range ops {
			op.apply(model)
		}
		changed = true
		model.Unlock()
	}

	if len(errs) == 0 {
		return changed, nil
	}
	if len(errs) == 1 {
		return changed, errs[0]
	}
	var errStrings []string
	for _, err := range errs {
		errStrings = append(errStrings, err.Error())
	}
	return changed, errors.New(StoreError{"MODEL_MULTIPLE", strings.Join(errStrings, "\n")})
}

// ModelError represents a modelling error
type ModelError struct {
	code    string
	message string
}

func (e ModelError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

func checkID(item Item) error {
	parts := strings.Split(item.ID, "/")
	if len(parts) < 2 {
		return errors.New(ModelError{"SHORT_ID",
			fmt.Sprintf("ID string is too short to represent type/id: %s", item.ID)})
	}
	if parts[len(parts)-2] != item.Type {
		return errors.New(ModelError{"NO_TYPE",
			fmt.Sprintf("ID string is does not contain item type: %s != %s", item.Type, parts[len(parts)-2])})
	}

	return nil
}

func parentType(model *Model, id string, itype string, ops []modelOperation) []modelOperation {
	parts := strings.Split(id, "/")
	var parent string

	if len(parts) > 3 {
		// /parentType/parentID/childType/childID, we want parentType
		parent = parts[len(parts)-4]
		ops = parentType(model, strings.Join(parts[:len(parts)-2], "/"), parent, ops)
	}
	_, ok := model.typeChildren[parent][itype]
	if !ok {
		ops = append(ops, addChild{parent, itype})
	}

	return ops
}

// ChildTypes returns the list of child types for a given parent type ("" for root types)
func ChildTypes(model *Model, parentType string) []string {
	model.RLock()
	types := childTypes(model, parentType)
	model.RUnlock()
	return types
}

func childTypes(model *Model, parentType string) []string {
	m1 := model.typeChildren[parentType]
	var types []string
	for k := range m1 {
		types = append(types, k)
	}
	return types
}
