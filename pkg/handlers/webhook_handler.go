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
	argsString := "("
	for _, a := range args {
		argParts := strings.Split(a, ":")
		key := strings.Trim(argParts[0], " ")
		value := strings.Trim(argParts[1], " ")
		tp := strings.Trim(argParts[2], " ")
		if tp == "" || tp == "String" {
			value = "\"" + value + "\""
		}
		argsString += " " + key + ": " + value
	}
	argsString += ")"
	return argsString
}

func query(path []string, args [][]string) string {
	if len(path) == 0 {
		return ""
	}
	return "{" + path[0] + buildArgs(args[0]) + query(path[1:], args[1:]) + "}"
}

func buildQuery(path []string, args [][]string) string {
	return "query " + query(path, args)
}

func buildMutation(path []string, args [][]string) string {
	return "mutation " + query(path, args)
}

type withFields interface {
	Fields() graphql.FieldDefinitionMap
}

func findReturnTypeForWebhook(schema *graphql.Schema, parent withFields, field string) (withFields, error) {
	fields := parent.Fields()
	fDef := fields[field]
	tp := fDef.Type.Name()
	fType, ok := schema.TypeMap()[tp]
	if !ok {
		return nil, errors.New("type not found: " + tp)
	}
	switch fType.(type) {
	case *graphql.Object, *graphql.Interface:
	}
}

func pathToPattern(
	schema *graphql.Schema,
	tp *graphql.Object,
	pathParts []string,
	pattern string,
) (fieldPath []string, args [][]string, err error) {
	patternParts := strings.Split(pattern, "/")
	if len(pathParts)-1 < len(patternParts) {
		err = errors.New("path does not match pattern")
		return
	}
	for i := 1; i < len(pathParts); i++ {
		if strings.HasPrefix(patternParts[i], "{") {
			err = errors.New("invalid webhook pattern")
			return
		}
		fieldPath = append(fieldPath, pathParts[i])
		args = append(args, []string{})
		for j := i + 1; j < len(pathParts) && strings.HasPrefix(patternParts[j], "{"); j++ {
			i = j
			argPattern := strings.Split(patternParts[j], ":")
			if len(argPattern) != 1 && len(argPattern) != 2 {
				err = errors.New("invalid argument pattern")
			}
			args[i] = append(
				args[i],
				strings.Join(append([]string{argPattern[j], pathParts[j]}, argPattern[1:]...), ":"),
			)
		}
	}
	return
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
		pathParts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/"), "/")
		op := strings.ToLower(pathParts[1])
		if len(pathParts) < 3 || (op != "query" && op != "mutation") || len(pathParts) == 0 {
			http.Error(rw, "Invalid webhook path, must be /webhook/<query|mutation>/<field>[/rest]", http.StatusNotFound)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(`{ "query": "` + op + `{` + pathParts[2] + `}" }`))
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
