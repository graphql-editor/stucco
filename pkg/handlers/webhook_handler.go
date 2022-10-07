package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-go/graphql"
	"k8s.io/klog"
)

func buildArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	argsString := ""
	for _, a := range args {
		argParts := strings.Split(a, ":")
		key := strings.Trim(argParts[0], " ")
		value := strings.Trim(argParts[1], " ")
		if len(argParts) > 2 {
			tp := strings.Trim(argParts[2], " ")
			if tp == "" || tp == "String" {
				value = "\"" + value + "\""
			}
		}
		argsString += " " + key + ": " + value
	}
	return "(" + strings.Trim(argsString, " ") + ")"
}

func query(path []string, args [][]string) string {
	if len(path) == 0 {
		return ""
	}
	return "{" + path[0] + buildArgs(args[0]) + query(path[1:], args[1:]) + "}"
}

func build(prefix string, path []string, args [][]string) string {
	return prefix + query(path, args)
}

type withFields interface {
	Fields() graphql.FieldDefinitionMap
	Name() string
}

func findReturnTypeForWebhook(cfg Config, parent withFields, field string) (graphql.Output, error) {
	fields := parent.Fields()
	fDef, ok := fields[field]
	var fType graphql.Output
	if ok {
		tp := fDef.Type.Name()
		fType, ok = cfg.Schema.TypeMap()[tp]
	}
	if !ok {
		return nil, errors.New("type for field not found: " + field)
	}
	return fType, nil
}

func typeWithFields(tp graphql.Output) (withFields, error) {
	switch wtp := tp.(type) {
	case *graphql.NonNull:
		return typeWithFields(wtp.OfType)
	case *graphql.List:
		return typeWithFields(wtp.OfType)
	case withFields:
		return wtp, nil
	default:
		return nil, errors.New("must be object or interface")
	}
}

func findWebhookConfig(cfg Config, tpName, fieldName string) router.WebhookConfig {
	tf := tpName + "." + fieldName
	for k, rcfg := range cfg.RouterConfig.Resolvers {
		if k == tf {
			var wcfg router.WebhookConfig
			if rcfg.Webhook != nil {
				wcfg = *rcfg.Webhook
			}
			return wcfg
		}
	}
	for k, icfg := range cfg.RouterConfig.Interfaces {
		if tpName == k {
			for ik, wcfg := range icfg.Webhooks {
				if ik == fieldName {
					return wcfg
				}
			}
			break
		}
	}
	return router.WebhookConfig{}
}

func pathToPattern(config Config, tp withFields, pathParts []string) (fieldPath []string, args [][]string, err error) {
	fieldPath = []string{pathParts[0]}
	args = [][]string{{}}
	if strings.HasPrefix(fieldPath[0], "{") {
		err = errors.New("invalid webhook pattern")
		return
	}
	pathParts = pathParts[1:]
	wcfg := findWebhookConfig(config, tp.Name(), fieldPath[0])
	var patternParts []string
	if len(wcfg.Pattern) > 0 {
		patternParts = strings.Split(strings.Trim(wcfg.Pattern, "/"), "/")
	}
	if len(pathParts) < len(patternParts) {
		err = errors.New("path does not match pattern")
		return
	}
	field, ok := tp.Fields()[fieldPath[0]]
	if !ok {
		return nil, nil, errors.New(fieldPath[0] + " does not exist on type " + tp.Name())
	}
	for i := 0; i < len(patternParts) && err == nil; i++ {
		argPattern := strings.Split(strings.Trim(strings.Trim(patternParts[i], "{"), "}"), ":")
		if len(argPattern) != 1 && len(argPattern) != 2 {
			err = errors.New("invalid argument pattern")
		}
		var arg *graphql.Argument
		for j := 0; j < len(field.Args) && arg == nil; j++ {
			if argPattern[0] == field.Args[j].Name() {
				arg = field.Args[j]
			}
		}
		if arg == nil {
			return nil, nil, errors.New("argument " + argPattern[0] + " does not exist on field " + fieldPath[0])
		}
		if len(argPattern) < 2 && arg.Type.Name() == "String" {
			argPattern = append(argPattern, "String")
		}
		args[0] = append(
			args[0],
			strings.Join(append([]string{argPattern[0], pathParts[0]}, argPattern[1:]...), ":"),
		)
		pathParts = pathParts[1:]
	}
	if err == nil && len(pathParts) > 0 {
		var o graphql.Output
		o, err = findReturnTypeForWebhook(config, tp, fieldPath[0])
		if err == nil {
			tp, err = typeWithFields(o)
			if err == nil {
				var npath []string
				var nargs [][]string
				npath, nargs, err = pathToPattern(config, tp, pathParts)
				if err == nil {
					fieldPath = append(fieldPath, npath...)
					args = append(args, nargs...)
				}
			}
		}
	}
	return
}

// CreateQuery from handler config and request
func CreateQuery(c Config, r *http.Request) (string, error) {
	pathParts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/"), "/")
	op := strings.ToLower(pathParts[1])
	if len(pathParts) < 3 {
		return "", errors.New("Invalid webhook path, must be /webhook/<query|mutation>/<field>[/rest]")
	}
	var tp *graphql.Object
	switch op {
	case "query":
		tp = c.Schema.QueryType()
	case "mutation":
		tp = c.Schema.MutationType()
	default:
		return "", errors.New(op + " is not a valid operation")
	}
	var q string
	path, args, err := pathToPattern(c, tp, pathParts[2:])
	if err == nil {
		q = build(op, path, args)
	}
	return q, err
}

func NewWebhookHandler(c Config, gqlHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		protocolData := map[string]interface{}{
			"method": r.Method,
			"url": map[string]string{
				"path":  r.URL.Path,
				"host":  r.URL.Host,
				"query": r.URL.RawQuery,
			},
			"host":          r.Host,
			"remoteAddress": r.RemoteAddr,
			"proto":         r.Proto,
			"headers":       r.Header.Clone(),
		}
		if r.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(rw, "could not read request body", http.StatusInternalServerError)
				return
			}
			r.Body.Close()
			if len(body) > 0 {
				protocolData["body"] = base64.StdEncoding.EncodeToString(body)
			}
		}
		q, err := CreateQuery(c, r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(q))
		r.Method = "POST"
		r.Header.Set("content-type", "application/json")
		r = r.WithContext(context.WithValue(r.Context(), router.ProtocolKey, protocolData))
		respProxy := webhookResponseWrapper{
			ResponseWriter: rw,
		}
		gqlHandler.ServeHTTP(&respProxy, r)
		status := respProxy.status
		if respProxy.status >= 200 && respProxy.status < 300 {
			var resp graphql.Result
			if err := json.Unmarshal(respProxy.Buffer.Bytes(), &resp); err != nil {
				status = http.StatusInternalServerError
			} else if resp.HasErrors() {
				status = http.StatusBadRequest
			}
		}
		rw.WriteHeader(status)
		if _, err := io.Copy(rw, &respProxy.Buffer); err != nil {
			klog.Error(err)
		}
	})
}
