package runtime

import (
	"fmt"
	"reflect"

	"github.com/rumlang/rum/parser"
)

// SetFn an function in parser function map
func (c *Context) SetFn(id parser.Identifier, v interface{}, adapters ...Adapter) {
	f := func(values ...interface{}) interface{} {
		args := values
		var err error
		for _, adapter := range adapters {
			args, err = adapter(args...)
			if err != nil {
				panic(fmt.Sprint("Error in adapter", values[0], err))
			}
		}

		vargs := []reflect.Value{}
		for _, arg := range args {
			vargs = append(vargs, reflect.ValueOf(arg))
		}

		result := reflect.ValueOf(v).Call(vargs)
		return result[0].Interface()
	}

	c.env[id] = parser.NewAny(f, nil)
}

//Invoke call a method from native value
func Invoke(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 {
		panic("Invalid arguments")
	}

	obj := ctx.MustEval(args[0]).Value()
	descriptor := args[1].String()
	descriptor = descriptor[5 : len(descriptor)-1]

	method := reflect.ValueOf(obj).MethodByName(descriptor)
	if method.IsValid() {
		vargs := []reflect.Value{}
		for _, arg := range args[2:] {
			vargs = append(vargs, reflect.ValueOf(ctx.MustEval(arg).Value()))
		}

		result := method.Call(vargs)
		return parser.NewAny(result[0].Interface(), nil)
	}

	if reflect.ValueOf(obj).Type().Kind() == reflect.Ptr {
		field := reflect.Indirect(reflect.ValueOf(obj)).FieldByName(descriptor)
		if field.IsValid() {
			return parser.NewAny(field.Interface(), nil)
		}
	} else {
		field := reflect.ValueOf(obj).FieldByName(descriptor)
		if field.IsValid() {
			return parser.NewAny(field.Interface(), nil)
		}
	}

	panic("Method or field not found: '" + descriptor +
		"' in type: " + reflect.ValueOf(obj).Type().String())
}

//New create a zero go value of given type
func New(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) != 1 {
		panic("Invalid arguments")
	}

	name := args[0].String()
	name = name[5 : len(name)-1]

	ret := reflect.New(ctx.typeRegistry[name])

	return parser.NewAny(ret, nil)
}

//Set assigns value to one golang value
func Set(ctx *Context, args ...parser.Value) parser.Value {
	if len(args) < 2 {
		panic("Invalid arguments")
	}

	obj := reflect.ValueOf(ctx.MustEval(args[0]).Value())
	val := reflect.ValueOf(ctx.MustEval(args[1]).Value())

	//obj = reflect.Indirect(obj).Elem()
	///obj = obj.Elem().Field(0)

	fmt.Println("**********>>", obj, obj.Type().String(),
		obj.Interface(), reflect.TypeOf(obj.Interface()))
	fmt.Println("**********>>", val, val.Type().String(),
		val.Interface(), reflect.TypeOf(val.Interface()))
	fmt.Println("**********>>", obj.Type().AssignableTo(val.Type()), obj.CanSet())

	obj.Set(val)

	return parser.NewAny(obj, nil)
}
