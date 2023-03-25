package typescript

import (
	"log"
	"reflect"
	"sort"
	"strings"
)

func NewModelConverter() *ModelConverter {
	return &ModelConverter{
		models:    make(map[string]interface{}),
		typeMap:   make(map[string]string),
		generated: make(map[string]bool),
	}
}

type ModelConverter struct {
	models    map[string]interface{}
	typeMap   map[string]string
	generated map[string]bool
}

func (c *ModelConverter) Add(model interface{}) {
	if model == nil {
		return
	}
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		model = reflect.ValueOf(model).Elem().Interface()
	}
	if reflect.TypeOf(model).Kind() != reflect.Struct {
		return
	}
	if reflect.TypeOf(model).Name() == "" {
		return
	}
	c.models[reflect.TypeOf(model).Name()] = model
}

func (c *ModelConverter) SetupTypeMap(typeMap map[string]string) {
	c.typeMap = typeMap
}

func (c *ModelConverter) ToString() string {
	modelStr := ""

	var names = make([]string, 0, len(c.models))
	for name := range c.models {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		modelStr += c.convertToInterface(c.models[name])
	}

	return modelStr
}

func (c *ModelConverter) convertToInterface(model any) string {
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		model = reflect.ValueOf(model).Elem().Interface()
	}

	nameOfModel := reflect.TypeOf(model).Name()
	if c.generated[nameOfModel] {
		return ""
	}

	outPutStr := ""
	modelStr := "export interface " + nameOfModel + " {\n"
	numOfField := reflect.TypeOf(model).NumField()
	for i := 0; i < numOfField; i++ {
		field := reflect.TypeOf(model).Field(i)
		fieldName := c.getFieldName(model, i)
		if fieldName == "" {
			continue
		}
		var fieldType string
		if _, ok := c.typeMap[field.Type.Name()]; ok {
			fieldType = c.typeMap[field.Type.Name()]
		} else if field.Type.Kind() == reflect.Slice {
			if field.Type.Elem().Kind() == reflect.Struct {
				outPutStr += c.convertToInterface(reflect.New(field.Type.Elem()).Interface())
				fieldType = field.Type.Elem().Name() + "[]"
			} else {
				fieldType = c.goTypeToTsType(field.Type.Elem().Name()) + "[]"
			}
		} else if field.Type.Kind() == reflect.Struct {
			fieldType = field.Type.Name()
			outPutStr += c.convertToInterface(reflect.New(field.Type).Interface())
		} else {
			fieldType = c.goTypeToTsType(field.Type.Name())
		}
		modelStr += "    " + fieldName + ": " + fieldType + ";\n"
	}

	modelStr += "}\n\n"
	outPutStr += modelStr
	c.generated[nameOfModel] = true
	return outPutStr
}

func (c *ModelConverter) goTypeToTsType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int":
		return "number"
	case "int64":
		return "number"
	case "uint":
		return "number"
	case "uint64":
		return "number"
	case "float64":
		return "number"
	case "float32":
		return "number"
	case "bool":
		return "boolean"
	default:
		log.Fatalln("unknown type", goType)
	}

	return ""
}

func (c *ModelConverter) getFieldName(model interface{}, index int) string {
	for _, tag := range []string{"json", "form", "uri"} {
		t := reflect.TypeOf(model).Field(index).Tag.Get(tag)
		if t != "" {
			var canOmit bool
			binding := reflect.TypeOf(model).Field(index).Tag.Get("binding")
			if strings.Contains(t, "omitempty") || !strings.Contains(binding, "required") {
				canOmit = true
			}
			t = strings.ReplaceAll(t, "omitempty", "")
			t = strings.ReplaceAll(t, ",", "")
			if canOmit {
				t += "?"
			}
			return t
		}
	}
	return ""
}
