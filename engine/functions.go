package engine

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/codingersid/legit-template/runtime"
)

// DefaultFunctions returns the default template functions
func DefaultFunctions() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"ltrim":     strings.TrimLeft,
		"rtrim":     strings.TrimRight,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split":     strings.Split,
		"join":      strings.Join,
		"repeat":    strings.Repeat,
		"substr":    substr,
		"length":    length,
		"nl2br":     nl2br,
		"ucfirst":   ucfirst,
		"lcfirst":   lcfirst,
		"slug":      slug,
		"limit":     limit,
		"wordLimit": wordLimit,

		// HTML functions
		"html":     template.HTMLEscapeString,
		"htmlAttr": template.HTMLEscaper,
		"js":       template.JSEscapeString,
		"url":      url.QueryEscape,
		"safeHTML": safeHTML,
		"safeJS":   safeJS,
		"safeURL":  safeURL,
		"safeCSS":  safeCSS,

		// Array/Slice functions
		"first":    first,
		"last":     last,
		"reverse":  reverse,
		"sortAsc":  sortAsc,
		"sortDesc": sortDesc,
		"unique":   unique,
		"pluck":    pluck,
		"where":    where,
		"groupBy":  groupBy,
		"chunk":    chunk,
		"flatten":  flatten,
		"slice":    sliceFunc,
		"append":   appendFunc,
		"prepend":  prependFunc,
		"merge":    mergeFunc,

		// Map functions
		"dict":   dict,
		"set":    setInMap,
		"unset":  unsetInMap,
		"keys":   keys,
		"values": values,
		"hasKey": hasKey,

		// Number functions
		"add":      add,
		"sub":      sub,
		"mul":      mul,
		"div":      div,
		"mod":      mod,
		"round":    round,
		"floor":    floor,
		"ceil":     ceil,
		"abs":      abs,
		"min":      minFunc,
		"max":      maxFunc,
		"currency": currency,
		"number":   number,
		"percent":  percent,

		// Date functions
		"date":      formatDate,
		"now":       time.Now,
		"ago":       ago,
		"diff":      dateDiff,
		"addDate":   addDate,
		"subDate":   subDate,
		"timestamp": timestamp,

		// Comparison functions
		"eq":  equal,
		"ne":  notEqual,
		"lt":  lessThan,
		"gt":  greaterThan,
		"lte": lessOrEqual,
		"gte": greaterOrEqual,
		"and": and,
		"or":  or,
		"not": not,

		// Utility functions
		"default":  defaultValue,
		"isset":    isset,
		"empty":    isEmpty,
		"dump":     dump,
		"json":     jsonEncode,
		"jsonDec":  jsonDecode,
		"seq":      seq,
		"until":    until,
		"index":    index,
		"printf":   fmt.Sprintf,
		"print":    fmt.Sprint,
		"coalesce": coalesce,
		"ternary":  ternary,
		"typeof":   typeof,
		"toInt":    toInt,
		"toFloat":  toFloat,
		"toString": toString,
		"toBool":   toBool,

		// Loop helper
		"newLoop": runtime.NewLoop,

		// Validation helpers
		"hasError": hasError,
		"getError": getError,

		// Class/Style helpers
		"classArray": classArray,
		"styleArray": styleArray,
	}
}

// String functions

func substr(s string, start int, length ...int) string {
	runes := []rune(s)
	if start < 0 {
		start = len(runes) + start
	}
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return ""
	}

	end := len(runes)
	if len(length) > 0 && length[0] >= 0 {
		end = start + length[0]
		if end > len(runes) {
			end = len(runes)
		}
	}

	return string(runes[start:end])
}

func length(v interface{}) int {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return len([]rune(rv.String()))
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len()
	default:
		return 0
	}
}

func nl2br(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
}

func ucfirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func lcfirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = []rune(strings.ToLower(string(runes[0])))[0]
	return string(runes)
}

func slug(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func limit(s string, n int, suffix ...string) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	end := "..."
	if len(suffix) > 0 {
		end = suffix[0]
	}
	return string(runes[:n]) + end
}

func wordLimit(s string, n int, suffix ...string) string {
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	end := "..."
	if len(suffix) > 0 {
		end = suffix[0]
	}
	return strings.Join(words[:n], " ") + end
}

// HTML safe functions

func safeHTML(s string) template.HTML {
	return template.HTML(s)
}

func safeJS(s string) template.JS {
	return template.JS(s)
}

func safeURL(s string) template.URL {
	return template.URL(s)
}

func safeCSS(s string) template.CSS {
	return template.CSS(s)
}

// Array/Slice functions

func first(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.Len() > 0 {
			return rv.Index(0).Interface()
		}
	}
	return nil
}

func last(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.Len() > 0 {
			return rv.Index(rv.Len() - 1).Interface()
		}
	}
	return nil
}

func reverse(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return v
	}

	length := rv.Len()
	result := reflect.MakeSlice(rv.Type(), length, length)
	for i := 0; i < length; i++ {
		result.Index(i).Set(rv.Index(length - 1 - i))
	}
	return result.Interface()
}

func sortAsc(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	sorted := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
	reflect.Copy(sorted, rv)

	sort.SliceStable(sorted.Interface(), func(i, j int) bool {
		return fmt.Sprint(sorted.Index(i).Interface()) < fmt.Sprint(sorted.Index(j).Interface())
	})

	return sorted.Interface()
}

func sortDesc(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	sorted := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len())
	reflect.Copy(sorted, rv)

	sort.SliceStable(sorted.Interface(), func(i, j int) bool {
		return fmt.Sprint(sorted.Index(i).Interface()) > fmt.Sprint(sorted.Index(j).Interface())
	})

	return sorted.Interface()
}

func unique(v interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	seen := make(map[interface{}]bool)
	result := reflect.MakeSlice(rv.Type(), 0, rv.Len())

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if !seen[item] {
			seen[item] = true
			result = reflect.Append(result, rv.Index(i))
		}
	}

	return result.Interface()
}

func pluck(v interface{}, key string) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make([]interface{}, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if item.Kind() == reflect.Map {
			if val := item.MapIndex(reflect.ValueOf(key)); val.IsValid() {
				result = append(result, val.Interface())
			}
		} else if item.Kind() == reflect.Struct {
			if field := item.FieldByName(key); field.IsValid() {
				result = append(result, field.Interface())
			}
		}
	}

	return result
}

func where(v interface{}, key string, value interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := reflect.MakeSlice(rv.Type(), 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		var itemVal interface{}

		if item.Kind() == reflect.Map {
			if val := item.MapIndex(reflect.ValueOf(key)); val.IsValid() {
				itemVal = val.Interface()
			}
		} else if item.Kind() == reflect.Struct {
			if field := item.FieldByName(key); field.IsValid() {
				itemVal = field.Interface()
			}
		}

		if itemVal == value {
			result = reflect.Append(result, item)
		}
	}

	return result.Interface()
}

func groupBy(v interface{}, key string) map[string][]interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	result := make(map[string][]interface{})
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		var groupKey string

		if item.Kind() == reflect.Map {
			if val := item.MapIndex(reflect.ValueOf(key)); val.IsValid() {
				groupKey = fmt.Sprint(val.Interface())
			}
		} else if item.Kind() == reflect.Struct {
			if field := item.FieldByName(key); field.IsValid() {
				groupKey = fmt.Sprint(field.Interface())
			}
		}

		result[groupKey] = append(result[groupKey], item.Interface())
	}

	return result
}

func chunk(v interface{}, size int) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice || size <= 0 {
		return v
	}

	length := rv.Len()
	numChunks := (length + size - 1) / size
	result := make([]interface{}, 0, numChunks)

	for i := 0; i < length; i += size {
		end := i + size
		if end > length {
			end = length
		}
		chunk := reflect.MakeSlice(rv.Type(), end-i, end-i)
		for j := i; j < end; j++ {
			chunk.Index(j - i).Set(rv.Index(j))
		}
		result = append(result, chunk.Interface())
	}

	return result
}

func flatten(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	result := make([]interface{}, 0)

	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return append(result, v)
	}

	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if item.Kind() == reflect.Slice || item.Kind() == reflect.Array {
			result = append(result, flatten(item.Interface())...)
		} else {
			result = append(result, item.Interface())
		}
	}

	return result
}

func sliceFunc(v interface{}, start int, end ...int) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return v
	}

	length := rv.Len()
	if start < 0 {
		start = length + start
	}
	if start < 0 {
		start = 0
	}

	endIdx := length
	if len(end) > 0 {
		endIdx = end[0]
		if endIdx < 0 {
			endIdx = length + endIdx
		}
	}
	if endIdx > length {
		endIdx = length
	}
	if start >= endIdx {
		return reflect.MakeSlice(rv.Type(), 0, 0).Interface()
	}

	return rv.Slice(start, endIdx).Interface()
}

func appendFunc(v interface{}, items ...interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	result := reflect.MakeSlice(rv.Type(), rv.Len(), rv.Len()+len(items))
	reflect.Copy(result, rv)

	for _, item := range items {
		result = reflect.Append(result, reflect.ValueOf(item))
	}

	return result.Interface()
}

func prependFunc(v interface{}, items ...interface{}) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return v
	}

	result := reflect.MakeSlice(rv.Type(), 0, len(items)+rv.Len())
	for _, item := range items {
		result = reflect.Append(result, reflect.ValueOf(item))
	}
	for i := 0; i < rv.Len(); i++ {
		result = reflect.Append(result, rv.Index(i))
	}

	return result.Interface()
}

func mergeFunc(maps ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		if m == nil {
			continue
		}
		rv := reflect.ValueOf(m)
		if rv.Kind() == reflect.Map {
			for _, key := range rv.MapKeys() {
				result[fmt.Sprint(key.Interface())] = rv.MapIndex(key).Interface()
			}
		}
	}
	return result
}

// Map functions

func dict(pairs ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i+1 < len(pairs); i += 2 {
		result[fmt.Sprint(pairs[i])] = pairs[i+1]
	}
	return result
}

func setInMap(m map[string]interface{}, key string, value interface{}) map[string]interface{} {
	if m == nil {
		m = make(map[string]interface{})
	}
	m[key] = value
	return m
}

func unsetInMap(m map[string]interface{}, key string) map[string]interface{} {
	if m != nil {
		delete(m, key)
	}
	return m
}

func keys(m interface{}) []string {
	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Map {
		return nil
	}

	result := make([]string, 0, rv.Len())
	for _, key := range rv.MapKeys() {
		result = append(result, fmt.Sprint(key.Interface()))
	}
	return result
}

func values(m interface{}) []interface{} {
	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Map {
		return nil
	}

	result := make([]interface{}, 0, rv.Len())
	for _, key := range rv.MapKeys() {
		result = append(result, rv.MapIndex(key).Interface())
	}
	return result
}

func hasKey(m interface{}, key string) bool {
	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Map {
		return false
	}
	return rv.MapIndex(reflect.ValueOf(key)).IsValid()
}

// Number functions

func add(a, b interface{}) interface{} {
	af := toFloat64(a)
	bf := toFloat64(b)
	return af + bf
}

func sub(a, b interface{}) interface{} {
	af := toFloat64(a)
	bf := toFloat64(b)
	return af - bf
}

func mul(a, b interface{}) interface{} {
	af := toFloat64(a)
	bf := toFloat64(b)
	return af * bf
}

func div(a, b interface{}) interface{} {
	af := toFloat64(a)
	bf := toFloat64(b)
	if bf == 0 {
		return 0
	}
	return af / bf
}

func mod(a, b interface{}) interface{} {
	ai := toInt64(a)
	bi := toInt64(b)
	if bi == 0 {
		return 0
	}
	return ai % bi
}

func round(n interface{}, precision ...int) float64 {
	nf := toFloat64(n)
	p := 0
	if len(precision) > 0 {
		p = precision[0]
	}
	mult := math.Pow(10, float64(p))
	return math.Round(nf*mult) / mult
}

func floor(n interface{}) float64 {
	return math.Floor(toFloat64(n))
}

func ceil(n interface{}) float64 {
	return math.Ceil(toFloat64(n))
}

func abs(n interface{}) float64 {
	return math.Abs(toFloat64(n))
}

func minFunc(values ...interface{}) interface{} {
	if len(values) == 0 {
		return nil
	}
	minVal := toFloat64(values[0])
	for _, v := range values[1:] {
		if f := toFloat64(v); f < minVal {
			minVal = f
		}
	}
	return minVal
}

func maxFunc(values ...interface{}) interface{} {
	if len(values) == 0 {
		return nil
	}
	maxVal := toFloat64(values[0])
	for _, v := range values[1:] {
		if f := toFloat64(v); f > maxVal {
			maxVal = f
		}
	}
	return maxVal
}

func currency(n interface{}, symbol ...string) string {
	nf := toFloat64(n)
	sym := "$"
	if len(symbol) > 0 {
		sym = symbol[0]
	}
	return fmt.Sprintf("%s%.2f", sym, nf)
}

func number(n interface{}, decimals ...int) string {
	nf := toFloat64(n)
	d := 0
	if len(decimals) > 0 {
		d = decimals[0]
	}
	return fmt.Sprintf("%.*f", d, nf)
}

func percent(n interface{}, decimals ...int) string {
	nf := toFloat64(n)
	d := 0
	if len(decimals) > 0 {
		d = decimals[0]
	}
	return fmt.Sprintf("%.*f%%", d, nf*100)
}

// Date functions

func formatDate(format string, t ...interface{}) string {
	var tm time.Time
	if len(t) > 0 {
		switch v := t[0].(type) {
		case time.Time:
			tm = v
		case string:
			tm, _ = time.Parse(time.RFC3339, v)
		case int64:
			tm = time.Unix(v, 0)
		default:
			tm = time.Now()
		}
	} else {
		tm = time.Now()
	}

	// Convert PHP date format to Go format
	format = convertDateFormat(format)
	return tm.Format(format)
}

func convertDateFormat(format string) string {
	replacements := map[string]string{
		"Y": "2006",
		"y": "06",
		"m": "01",
		"n": "1",
		"d": "02",
		"j": "2",
		"H": "15",
		"h": "03",
		"i": "04",
		"s": "05",
		"A": "PM",
		"a": "pm",
		"M": "Jan",
		"F": "January",
		"D": "Mon",
		"l": "Monday",
	}

	for php, goFmt := range replacements {
		format = strings.ReplaceAll(format, php, goFmt)
	}
	return format
}

func ago(t interface{}) string {
	var tm time.Time
	switch v := t.(type) {
	case time.Time:
		tm = v
	case string:
		tm, _ = time.Parse(time.RFC3339, v)
	case int64:
		tm = time.Unix(v, 0)
	default:
		return ""
	}

	diff := time.Since(tm)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 30*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

func dateDiff(t1, t2 interface{}) time.Duration {
	parse := func(t interface{}) time.Time {
		switch v := t.(type) {
		case time.Time:
			return v
		case string:
			tm, _ := time.Parse(time.RFC3339, v)
			return tm
		case int64:
			return time.Unix(v, 0)
		default:
			return time.Time{}
		}
	}
	return parse(t2).Sub(parse(t1))
}

func addDate(t interface{}, years, months, days int) time.Time {
	var tm time.Time
	switch v := t.(type) {
	case time.Time:
		tm = v
	case string:
		tm, _ = time.Parse(time.RFC3339, v)
	case int64:
		tm = time.Unix(v, 0)
	default:
		tm = time.Now()
	}
	return tm.AddDate(years, months, days)
}

func subDate(t interface{}, years, months, days int) time.Time {
	return addDate(t, -years, -months, -days)
}

func timestamp(t ...interface{}) int64 {
	if len(t) > 0 {
		switch v := t[0].(type) {
		case time.Time:
			return v.Unix()
		case string:
			tm, _ := time.Parse(time.RFC3339, v)
			return tm.Unix()
		}
	}
	return time.Now().Unix()
}

// Comparison functions

func equal(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

func notEqual(a, b interface{}) bool {
	return !reflect.DeepEqual(a, b)
}

func lessThan(a, b interface{}) bool {
	return toFloat64(a) < toFloat64(b)
}

func greaterThan(a, b interface{}) bool {
	return toFloat64(a) > toFloat64(b)
}

func lessOrEqual(a, b interface{}) bool {
	return toFloat64(a) <= toFloat64(b)
}

func greaterOrEqual(a, b interface{}) bool {
	return toFloat64(a) >= toFloat64(b)
}

func and(values ...interface{}) bool {
	for _, v := range values {
		if !toBoolValue(v) {
			return false
		}
	}
	return true
}

func or(values ...interface{}) bool {
	for _, v := range values {
		if toBoolValue(v) {
			return true
		}
	}
	return false
}

func not(v interface{}) bool {
	return !toBoolValue(v)
}

// Utility functions

func defaultValue(value, def interface{}) interface{} {
	if isEmpty(value) {
		return def
	}
	return value
}

func isset(v interface{}) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return !rv.IsNil()
	case reflect.Invalid:
		return false
	}
	return true
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

func dump(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func jsonEncode(v interface{}) template.JS {
	b, _ := json.Marshal(v)
	return template.JS(b)
}

func jsonDecode(s string) interface{} {
	var result interface{}
	json.Unmarshal([]byte(s), &result)
	return result
}

func seq(start, end interface{}) []int {
	s := int(toInt64(start))
	e := int(toInt64(end))
	if s > e {
		result := make([]int, s-e+1)
		for i := range result {
			result[i] = s - i
		}
		return result
	}
	result := make([]int, e-s+1)
	for i := range result {
		result[i] = s + i
	}
	return result
}

func until(n interface{}) []int {
	count := int(toInt64(n))
	if count <= 0 {
		return nil
	}
	result := make([]int, count)
	for i := range result {
		result[i] = i
	}
	return result
}

func index(v interface{}, key interface{}) interface{} {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		i := int(toInt64(key))
		if i >= 0 && i < rv.Len() {
			return rv.Index(i).Interface()
		}
	case reflect.Map:
		if val := rv.MapIndex(reflect.ValueOf(key)); val.IsValid() {
			return val.Interface()
		}
	}
	return nil
}

func coalesce(values ...interface{}) interface{} {
	for _, v := range values {
		if !isEmpty(v) {
			return v
		}
	}
	return nil
}

func ternary(cond bool, trueVal, falseVal interface{}) interface{} {
	if cond {
		return trueVal
	}
	return falseVal
}

func typeof(v interface{}) string {
	return reflect.TypeOf(v).String()
}

func toInt(v interface{}) int {
	return int(toInt64(v))
}

func toFloat(v interface{}) float64 {
	return toFloat64(v)
}

func toString(v interface{}) string {
	return fmt.Sprint(v)
}

func toBool(v interface{}) bool {
	return toBoolValue(v)
}

// Validation helpers

func hasError(errors interface{}, field string) bool {
	if errors == nil {
		return false
	}
	rv := reflect.ValueOf(errors)
	if rv.Kind() == reflect.Map {
		if val := rv.MapIndex(reflect.ValueOf(field)); val.IsValid() {
			if arr := val.Interface(); arr != nil {
				if slice, ok := arr.([]string); ok {
					return len(slice) > 0
				}
			}
		}
	}
	return false
}

func getError(errors interface{}, field string) string {
	if errors == nil {
		return ""
	}
	rv := reflect.ValueOf(errors)
	if rv.Kind() == reflect.Map {
		if val := rv.MapIndex(reflect.ValueOf(field)); val.IsValid() {
			if arr := val.Interface(); arr != nil {
				if slice, ok := arr.([]string); ok && len(slice) > 0 {
					return slice[0]
				}
			}
		}
	}
	return ""
}

// Class/Style helpers

func classArray(classes interface{}) string {
	rv := reflect.ValueOf(classes)
	if rv.Kind() != reflect.Slice {
		return ""
	}

	var result []string
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return strings.Join(result, " ")
}

func styleArray(styles interface{}) string {
	rv := reflect.ValueOf(styles)
	if rv.Kind() != reflect.Map {
		return ""
	}

	var result []string
	for _, key := range rv.MapKeys() {
		val := rv.MapIndex(key)
		if toBoolValue(val.Interface()) {
			result = append(result, fmt.Sprint(key.Interface()))
		}
	}
	return strings.Join(result, "; ")
}

// Helper conversion functions

func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int8:
		return float64(n)
	case int16:
		return float64(n)
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case uint8:
		return float64(n)
	case uint16:
		return float64(n)
	case uint32:
		return float64(n)
	case uint64:
		return float64(n)
	case float32:
		return float64(n)
	case float64:
		return n
	case string:
		f := 0.0
		fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint:
		return int64(n)
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)
	case float32:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		i := int64(0)
		fmt.Sscanf(n, "%d", &i)
		return i
	default:
		return 0
	}
}

func toBoolValue(v interface{}) bool {
	if v == nil {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(b).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(b).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(b).Float() != 0
	case string:
		return b != "" && b != "0" && b != "false"
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array, reflect.Map:
			return rv.Len() > 0
		case reflect.Ptr, reflect.Interface:
			return !rv.IsNil()
		}
		return true
	}
}
