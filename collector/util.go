package collector

import "supos.ai/data-lake/external/tsdb-proxy/common/model"

// NewMetaObject new MetaObject
func NewMetaObject(name string, values *model.ValueSequnce) *model.MetaObject {
	obj := &model.MetaObject{Name: name, TemplateName: "LinkObject"}

	internelValues := &model.ValueSequnce{}

	onlineVal := &model.NamedValue{Name: "online", Value: &model.Value{Kind: &model.Value_PrimitiveValue{PrimitiveValue: &model.PrimitiveValue{Value: &model.PrimitiveValue_BoolValue{BoolValue: true}}}}}
	internelValues.Value = append(internelValues.Value, onlineVal)
	internalPros := ValueSequnce2MetaProperty(internelValues)
	for _, val := range internalPros {
		sp := &model.PropertyOrObject{Filed: &model.PropertyOrObject_Prop{Prop: val}}
		obj.Field = append(obj.Field, sp)
	}

	pros := ValueSequnce2MetaProperty(values)
	for _, val := range pros {
		sp := &model.PropertyOrObject{Filed: &model.PropertyOrObject_Prop{Prop: val}}
		obj.Field = append(obj.Field, sp)
	}

	return obj
}

// ValueSequnce2MetaProperty ValueSequnce 2 meta property
func ValueSequnce2MetaProperty(values *model.ValueSequnce) []*model.MetaProperty {
	properies := []*model.MetaProperty{}
	for _, val := range values.Value {
		pro := Value2MetaProperty(val)
		if pro != nil {
			properies = append(properies, pro)
		}
	}

	return properies
}

// Value2MetaProperty NamedValue 2 meta property
func Value2MetaProperty(val *model.NamedValue) *model.MetaProperty {
	pro := &model.MetaProperty{Name: val.GetName()}

	primiteType := getValueType(val.GetValue())
	if primiteType == "" {
		return nil
	}

	pro.PrimitiveType = primiteType
	pro.DefaultValue = val.GetValue()

	return pro
}

func getValueType(val *model.Value) string {
	kindVal := val.GetKind()

	primiteValue, ok := kindVal.(*model.Value_PrimitiveValue)
	if ok {
		return getPrimiteType(primiteValue.PrimitiveValue)
	}

	primiteVQTValue, ok := kindVal.(*model.Value_PrimitiveValueWithQT)
	if ok {
		return getPrimiteType(primiteVQTValue.PrimitiveValueWithQT.GetValue())
	}

	_, ok = kindVal.(*model.Value_TableValue)
	if ok {
		return "DataTable"
	}

	return ""
}

func getPrimiteType(primitiveVal *model.PrimitiveValue) string {
	value := primitiveVal.GetValue()

	_, ok := value.(*model.PrimitiveValue_I32Value)
	if ok {
		return "Integer"
	}

	_, ok = value.(*model.PrimitiveValue_FltValue)
	if ok {
		return "Float"
	}

	_, ok = value.(*model.PrimitiveValue_DblValue)
	if ok {
		return "Double"
	}

	_, ok = value.(*model.PrimitiveValue_BoolValue)
	if ok {
		return "Boolean"
	}
	_, ok = value.(*model.PrimitiveValue_StrValue)
	if ok {
		return "String"
	}
	_, ok = value.(*model.PrimitiveValue_I64Value)
	if ok {
		return "Long"
	}
	_, ok = value.(*model.PrimitiveValue_Ui32Value)
	if ok {
		return "Long"
	}
	_, ok = value.(*model.PrimitiveValue_Ui64Value)
	if ok {
		return "Long"
	}

	return ""
}
