package reflect

import (
	"errors"
	"reflect"
)

type FuncInfo struct {
	Name        string
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	typ := reflect.TypeOf(entity)

	if typ.Kind() != reflect.Ptr && typ.Kind() != reflect.Struct {
		return nil, errors.New("非法类型")
	}

	numMethod := typ.NumMethod()

	res := make(map[string]FuncInfo)

	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func

		// 入参
		// 先获取有多少个入参
		numIn := fn.Type().NumIn()

		input := make([]reflect.Type, 0, numIn)
		// 调用方法需要使用
		inputValue := make([]reflect.Value, 0, numIn)

		// go中 func (*user)Age = func Age(*user)
		// 第0个等于函数本身，需要单独附加上去
		// 第0个一定是只想entity本身
		inputValue = append(inputValue, reflect.ValueOf(entity))
		input = append(input, reflect.TypeOf(entity))

		// 循环加入，从第一个开始
		// 第一个开始才是函数的输入参数
		for j := 1; j < numIn; j++ {
			input = append(input, fn.Type().In(j))
			// 调用方法用0值发起调用
			inputValue = append(inputValue, reflect.Zero(fn.Type().In(j)))
		}

		// 出参
		numOut := fn.Type().NumOut()

		outPut := make([]reflect.Type, 0, numOut)
		// 循环加入
		for j := 0; j < numOut; j++ {
			outPut = append(outPut, fn.Type().Out(j))
		}

		resValues := fn.Call(inputValue)
		result := make([]any, 0, len(resValues))
		// 获取返回的值
		for _, v := range resValues {
			result = append(result, v.Interface())
		}

		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  input,
			OutputTypes: outPut,
			Result:      result,
		}
	}
	return res, nil
}
