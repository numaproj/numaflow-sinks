package expr

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Masterminds/sprig/v3"
	"github.com/antonmedv/expr"
)

var sprigFuncMap = sprig.GenericFuncMap()

const JsonRoot = "payload"

func EvalExpression(expression string, msg []byte) (interface{}, error) {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(msg, &jsonMap)
	if err != nil {
		return nil, err
	}
	msgMap := map[string]interface{}{
		JsonRoot: jsonMap,
	}
	env := GetFuncMap(msgMap)
	result, err := expr.Eval(expression, env)
	if err != nil {
		return nil, fmt.Errorf("unable to evaluate expression '%s': %s", expression, err)
	}
	return result, nil
}

func GetFuncMap(m map[string]interface{}) map[string]interface{} {
	env := Expand(m)
	env["sprig"] = sprigFuncMap
	env["json"] = _json
	env["int"] = _int
	env["string"] = _string
	return env
}

func _int(v interface{}) int {
	switch w := v.(type) {
	case []byte:
		i, err := strconv.Atoi(string(w))
		if err != nil {
			panic(fmt.Errorf("cannot convert %q an int", v))
		}
		return i
	case string:
		i, err := strconv.Atoi(w)
		if err != nil {
			panic(fmt.Errorf("cannot convert %q to int", v))
		}
		return i
	case float64:
		return int(w)
	case int:
		return w
	default:
		panic(fmt.Errorf("cannot convert %q to int", v))
	}
}

func _string(v interface{}) string {
	switch w := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(w)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func _json(v interface{}) map[string]interface{} {
	x := make(map[string]interface{})
	switch w := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(w, &x); err != nil {
			panic(fmt.Errorf("cannot convert %q to object: %v", v, err))
		}
		return x
	case string:
		if err := json.Unmarshal([]byte(w), &x); err != nil {
			panic(fmt.Errorf("cannot convert %q to object: %v", v, err))
		}
		return x
	default:
		panic("unknown type")
	}
}
