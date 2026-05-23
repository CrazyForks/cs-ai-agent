package ginx

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

var jsonResultType = reflect.TypeOf((*web.JsonResult)(nil))

func HandleController(group *gin.RouterGroup, relativePath string, prototype any, handlers ...gin.HandlerFunc) {
	router := group.Group(relativePath, handlers...)
	t := reflect.TypeOf(prototype)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("ginx.HandleController requires a pointer to a controller struct")
	}
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		httpMethods, path, ok := parseAction(method)
		if !ok {
			continue
		}
		handler := buildHandler(t, method)
		for _, httpMethod := range httpMethods {
			router.Handle(httpMethod, path, handler)
		}
	}
}

func parseAction(method reflect.Method) ([]string, string, bool) {
	prefixes := []struct {
		name    string
		methods []string
	}{
		{"Any", []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions}},
		{"Get", []string{http.MethodGet}},
		{"Post", []string{http.MethodPost}},
		{"Put", []string{http.MethodPut}},
		{"Delete", []string{http.MethodDelete}},
	}
	for _, prefix := range prefixes {
		if !strings.HasPrefix(method.Name, prefix.name) {
			continue
		}
		suffix := strings.TrimPrefix(method.Name, prefix.name)
		return prefix.methods, actionPath(suffix, method.Type.NumIn()-1), true
	}
	return nil, "", false
}

func actionPath(suffix string, argCount int) string {
	if suffix == "" {
		return "/"
	}
	if suffix == "By" && argCount == 1 {
		return "/:id"
	}
	if strings.HasSuffix(suffix, "By") && argCount == 1 {
		base := strings.TrimSuffix(suffix, "By")
		return "/" + actionSegmentPath(base) + "/:id"
	}
	return "/" + actionSegmentPath(suffix)
}

func actionSegmentPath(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	for i, part := range parts {
		parts[i] = camelToPath(part)
	}
	return strings.Join(parts, "_")
}

func camelToPath(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('/')
			}
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}

func buildHandler(controllerType reflect.Type, method reflect.Method) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		controller := reflect.New(controllerType.Elem())
		if field := controller.Elem().FieldByName("Ctx"); field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(ctx))
		}

		args := []reflect.Value{controller}
		for i := 1; i < method.Type.NumIn(); i++ {
			argType := method.Type.In(i)
			raw := ctx.Param("id")
			value, ok := convertPathArg(raw, argType)
			if !ok {
				ctx.JSON(http.StatusBadRequest, web.JsonErrorMsg("路径参数错误"))
				return
			}
			args = append(args, value)
		}

		results := method.Func.Call(args)
		if len(results) == 0 || results[0].IsNil() {
			return
		}
		if result, ok := results[0].Interface().(*web.JsonResult); ok {
			ctx.JSON(http.StatusOK, result)
		}
	}
}

func convertPathArg(raw string, t reflect.Type) (reflect.Value, bool) {
	switch t.Kind() {
	case reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(v), true
	case reflect.String:
		return reflect.ValueOf(raw), true
	default:
		return reflect.Zero(t), false
	}
}
