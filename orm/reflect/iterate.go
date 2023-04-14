package reflect

import "reflect"

func IterateArrayOrSlice(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		ele := val.Index(i)
		res = append(res, ele.Interface())
	}

	return res, nil
}

func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)
	resKey := make([]any, 0)
	resVal := make([]any, 0)

	// 两种用法都一样
	//itr := val.MapRange()
	//for itr.Next() {
	//	resKey = append(resKey, itr.Key().Interface())
	//	resVal = append(resVal, itr.Value().Interface())
	//}

	for _, key := range val.MapKeys() {
		v := val.MapIndex(key)
		resKey = append(resKey, key)
		resVal = append(resVal, v.Interface())
	}
	return resKey, resVal, nil

}
