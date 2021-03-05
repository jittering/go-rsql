package rsql

import (
	"log"
	"reflect"
	"strings"
)

// StructField :
type StructField struct {
	Name string
	Tag  *StructTag
	Type reflect.Type
}

// StructTag :
type StructTag struct {
	name   string
	values map[string]string
}

// Struct :
type Struct struct {
	Fields []*StructField
	Names  map[string]*StructField
}

func createJSONTag(fv reflect.StructField, tagVal string) *StructTag {
	t := new(StructTag)
	t.values = make(map[string]string)
	if strings.ContainsRune(tagVal, ',') {
		parts := strings.Split(tagVal, ",")
		if parts[0] != "" {
			t.name = parts[0]
		} else if parts[0] == "" {
			t.name = fv.Name
		}
	} else if tagVal == "-" || tagVal == "" {
		t.name = fv.Name
	} else {
		t.name = tagVal
	}
	t.values["filter"] = ""
	t.values["sort"] = ""
	t.values["allow"] = strings.Join(allowAll, "|")
	return t
}

func NewTag(fv reflect.StructField) *StructTag {
	tagVal := fv.Tag.Get("rsql")
	if tagVal == "" {
		// look for json tag instead
		tagVal = fv.Tag.Get("json")
		if tagVal != "" {
			return createJSONTag(fv, tagVal)
		}
	}

	paths := strings.Split(tagVal, ",")
	t := new(StructTag)
	t.name = paths[0]
	t.values = make(map[string]string)
	for _, v := range paths[1:] {
		p := strings.SplitN(v, "=", 2)
		t.values[p[0]] = ""
		if len(p) > 1 {
			t.values[p[0]] = p[1]
		}
	}
	return t
}

func (t StructTag) Lookup(key string) (value string, ok bool) {
	value, ok = t.values[key]
	return
}

func getCodec(t reflect.Type) *Struct {
	fields := make([]*StructField, 0)
	codec := new(Struct)
	for i := 0; i < t.NumField(); i++ {
		fv := t.Field(i)
		log.Println(fv)

		tag := NewTag(fv)
		f := new(StructField)
		f.Name = fv.Name
		f.Tag = tag
		if tag.name != "" {
			f.Name = tag.name
		}
		f.Type = fv.Type

		fields = append(fields, f)
	}

	codec.Fields = fields
	codec.Names = make(map[string]*StructField)
	for _, f := range fields {
		codec.Names[f.Name] = f
	}
	return codec
}
