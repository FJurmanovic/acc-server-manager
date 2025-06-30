package common

import (
	"acc-server-manager/local/utl/logging"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RouteGroups struct {
	Api          fiber.Router
	Auth         fiber.Router
	Server       fiber.Router
	Config       fiber.Router
	Lookup       fiber.Router
	StateHistory fiber.Router
	Membership   fiber.Router
}

func CheckError(err error) {
	if err != nil {
		logging.Error("Error occured. %v", err)
	}
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func Find[T any](lst *[]T, callback func(item *T) bool) *T {
	for _, item := range *lst {
		if callback(&item) {
			return &item
		}
	}
	return nil
}

func IndentJson(body []byte) ([]byte, error) {
	newBody := new([]byte)
	unmarshaledBody := bytes.NewBuffer(*newBody)
	err := json.Indent(unmarshaledBody, body, "", "  ")
	if err != nil {
		return nil, err
	}
	return unmarshaledBody.Bytes(), nil
}

// ParseQueryFilter parses query parameters into a filter struct using reflection.
// It supports various field types and uses struct tags to determine parsing behavior.
// Supported tags:
// - `query:"field_name"` - specifies the query parameter name
// - `param:"param_name"` - specifies the path parameter name
// - `time_format:"format"` - specifies the time format for parsing dates (default: RFC3339)
func ParseQueryFilter(c *fiber.Ctx, filter interface{}) error {
	val := reflect.ValueOf(filter)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("filter must be a non-nil pointer")
	}

	elem := val.Elem()
	typ := elem.Type()

	// Process all fields including embedded structs
	var processFields func(reflect.Value, reflect.Type) error
	processFields = func(val reflect.Value, typ reflect.Type) error {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)

			// Handle embedded structs recursively
			if fieldType.Anonymous {
				if err := processFields(field, fieldType.Type); err != nil {
					return err
				}
				continue
			}

			// Skip if field cannot be set
			if !field.CanSet() {
				continue
			}

			// Check for param tag first (path parameters)
			if paramName := fieldType.Tag.Get("param"); paramName != "" {
				if err := parsePathParam(c, field, paramName); err != nil {
					return fmt.Errorf("error parsing path parameter %s: %v", paramName, err)
				}
				continue
			}

			// Then check for query tag
			queryName := fieldType.Tag.Get("query")
			if queryName == "" {
				queryName = ToSnakeCase(fieldType.Name) // Default to snake_case of field name
			}

			queryVal := c.Query(queryName)
			if queryVal == "" {
				continue // Skip empty values
			}

			if err := parseValue(field, queryVal, fieldType.Tag); err != nil {
				return fmt.Errorf("error parsing query parameter %s: %v", queryName, err)
			}
		}
		return nil
	}

	return processFields(elem, typ)
}

func parsePathParam(c *fiber.Ctx, field reflect.Value, paramName string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := c.ParamsInt(paramName)
		if err != nil {
			if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
				return nil
			}
			return err
		}
		field.SetInt(int64(val))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := c.ParamsInt(paramName)
		if err != nil {
			if strings.Contains(err.Error(), "strconv.Atoi: parsing \"\": invalid syntax") {
				return nil
			}
			return err
		}
		field.SetUint(uint64(val))
	case reflect.String:
		field.SetString(c.Params(paramName))
	default:
		return fmt.Errorf("unsupported path parameter type: %v", field.Kind())
	}
	return nil
}

func parseValue(field reflect.Value, value string, tag reflect.StructTag) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(val)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)

	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			format := tag.Get("time_format")
			if format == "" {
				format = time.RFC3339
			}
			t, err := time.Parse(format, value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(t))
		} else {
			return fmt.Errorf("unsupported struct type: %v", field.Type())
		}

	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}

	return nil
}
