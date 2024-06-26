package interpreter

type ValueType string

const (
	VoidValue     ValueType = "Void"
	NullValue     ValueType = "Null"
	ArrayValue    ValueType = "Array"
	BooleanValue  ValueType = "Boolean"
	IntegerValue  ValueType = "Integer"
	FloatingValue ValueType = "Floating"
	StringValue   ValueType = "String"
)

// RuntimeValue

type IRuntimeValue interface {
	GetType() ValueType
}

type RuntimeValue struct {
	valueType ValueType
}

func NewRuntimeValue(valueType ValueType) *RuntimeValue {
	return &RuntimeValue{valueType: valueType}
}

func (runtimeValue *RuntimeValue) GetType() ValueType {
	return runtimeValue.valueType
}

// VoidValue

func NewVoidRuntimeValue() *RuntimeValue {
	return &RuntimeValue{valueType: VoidValue}
}

// NullValue

func NewNullRuntimeValue() *RuntimeValue {
	return &RuntimeValue{valueType: NullValue}
}

// ArrayRuntimeValue

type IArrayRuntimeValue interface {
	IRuntimeValue
	GetKeys() []IRuntimeValue
	GetElements() map[IRuntimeValue]IRuntimeValue
	GetElement(key IRuntimeValue) (IRuntimeValue, bool)
	SetElement(key IRuntimeValue, value IRuntimeValue)
}

type ArrayRuntimeValue struct {
	runtimeValue *RuntimeValue
	keys         []IRuntimeValue
	elements     map[IRuntimeValue]IRuntimeValue
}

func NewArrayRuntimeValue() *ArrayRuntimeValue {
	return &ArrayRuntimeValue{
		runtimeValue: NewRuntimeValue(ArrayValue),
		keys:         []IRuntimeValue{},
		elements:     map[IRuntimeValue]IRuntimeValue{},
	}
}

func NewArrayRuntimeValueFromMap(elements map[IRuntimeValue]IRuntimeValue) *ArrayRuntimeValue {
	keys := []IRuntimeValue{}
	for key := range elements {
		keys = append(keys, key)
	}
	return &ArrayRuntimeValue{
		runtimeValue: NewRuntimeValue(ArrayValue),
		keys:         keys,
		elements:     elements,
	}
}

func (runtimeValue *ArrayRuntimeValue) GetType() ValueType {
	return runtimeValue.runtimeValue.valueType
}

func (runtimeValue *ArrayRuntimeValue) SetElement(key IRuntimeValue, value IRuntimeValue) {
	existingKey, exists := runtimeValue.findKey(key)
	if !exists {
		runtimeValue.keys = append(runtimeValue.keys, key)
		runtimeValue.elements[key] = value
	} else {
		runtimeValue.elements[existingKey] = value
	}
}

func (runtimeValue *ArrayRuntimeValue) findKey(key IRuntimeValue) (IRuntimeValue, bool) {
	for k := range runtimeValue.elements {
		if k.GetType() != key.GetType() {
			continue
		}
		boolean, err := compare(key, "===", k)
		if err != nil {
			return NewVoidRuntimeValue(), false
		}
		if runtimeValToBoolRuntimeVal(boolean).GetValue() {
			return k, true
		}
	}
	return NewVoidRuntimeValue(), false
}

func (runtimeValue *ArrayRuntimeValue) GetKeys() []IRuntimeValue {
	return runtimeValue.keys
}

func (runtimeValue *ArrayRuntimeValue) GetElements() map[IRuntimeValue]IRuntimeValue {
	return runtimeValue.elements
}

func (runtimeValue *ArrayRuntimeValue) GetElement(key IRuntimeValue) (IRuntimeValue, bool) {
	key, found := runtimeValue.findKey(key)
	if !found {
		return NewVoidRuntimeValue(), false
	}
	return runtimeValue.elements[key], true
}

func runtimeValToArrayRuntimeVal(runtimeValue IRuntimeValue) IArrayRuntimeValue {
	var i interface{} = runtimeValue
	return i.(IArrayRuntimeValue)
}

// BooleanRuntimeValue

type IBooleanRuntimeValue interface {
	IRuntimeValue
	GetValue() bool
}

type BooleanRuntimeValue struct {
	runtimeValue *RuntimeValue
	value        bool
}

func NewBooleanRuntimeValue(value bool) *BooleanRuntimeValue {
	return &BooleanRuntimeValue{runtimeValue: NewRuntimeValue(BooleanValue), value: value}
}

func (runtimeValue *BooleanRuntimeValue) GetType() ValueType {
	return runtimeValue.runtimeValue.valueType
}

func (runtimeValue *BooleanRuntimeValue) GetValue() bool {
	return runtimeValue.value
}

func runtimeValToBoolRuntimeVal(runtimeValue IRuntimeValue) IBooleanRuntimeValue {
	var i interface{} = runtimeValue
	return i.(IBooleanRuntimeValue)
}

// IntegerRuntimeValue

type IIntegerRuntimeValue interface {
	IRuntimeValue
	GetValue() int64
}

type IntegerRuntimeValue struct {
	runtimeValue *RuntimeValue
	value        int64
}

func NewIntegerRuntimeValue(value int64) *IntegerRuntimeValue {
	return &IntegerRuntimeValue{runtimeValue: NewRuntimeValue(IntegerValue), value: value}
}

func (runtimeValue *IntegerRuntimeValue) GetType() ValueType {
	return runtimeValue.runtimeValue.valueType
}

func (runtimeValue *IntegerRuntimeValue) GetValue() int64 {
	return runtimeValue.value
}

func runtimeValToIntRuntimeVal(runtimeValue IRuntimeValue) IIntegerRuntimeValue {
	var i interface{} = runtimeValue
	return i.(IIntegerRuntimeValue)
}

// FloatingRuntimeValue

type IFloatingRuntimeValue interface {
	IRuntimeValue
	GetValue() float64
}

type FloatingRuntimeValue struct {
	runtimeValue *RuntimeValue
	value        float64
}

func NewFloatingRuntimeValue(value float64) *FloatingRuntimeValue {
	return &FloatingRuntimeValue{runtimeValue: NewRuntimeValue(FloatingValue), value: value}
}

func (runtimeValue *FloatingRuntimeValue) GetType() ValueType {
	return runtimeValue.runtimeValue.valueType
}

func (runtimeValue *FloatingRuntimeValue) GetValue() float64 {
	return runtimeValue.value
}

func runtimeValToFloatRuntimeVal(runtimeValue IRuntimeValue) IFloatingRuntimeValue {
	var i interface{} = runtimeValue
	return i.(IFloatingRuntimeValue)
}

// StringRuntimeValue

type IStringRuntimeValue interface {
	IRuntimeValue
	GetValue() string
}

type StringRuntimeValue struct {
	runtimeValue *RuntimeValue
	value        string
}

func NewStringRuntimeValue(value string) *StringRuntimeValue {
	return &StringRuntimeValue{runtimeValue: NewRuntimeValue(StringValue), value: value}
}

func (runtimeValue *StringRuntimeValue) GetType() ValueType {
	return runtimeValue.runtimeValue.valueType
}

func (runtimeValue *StringRuntimeValue) GetValue() string {
	return runtimeValue.value
}

func runtimeValToStrRuntimeVal(runtimeValue IRuntimeValue) IStringRuntimeValue {
	var i interface{} = runtimeValue
	return i.(IStringRuntimeValue)
}
