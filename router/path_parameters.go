package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type contextKey string

const ContextParamsKey contextKey = "reuqest_path_params"

type ParamsKey string

type ParameterMap = map[ParamsKey]string

func ReadParameters(r *http.Request) ParameterMap {
	ctx := r.Context()
	if parameterMap := ctx.Value(ContextParamsKey); nil == parameterMap {
		return make(ParameterMap)
	} else {
		return parameterMap.(ParameterMap)
	}
}

type ParameterDescription interface {
	getParameterName() string
	validate(parameterValue string) error
}

type parameterDescrition struct {
	name              string
	validatorFunction func(string) error
}

func (pD *parameterDescrition) getParameterName() string {
	return pD.name
}

func (pD *parameterDescrition) validate(parameterValue string) error {
	return pD.validatorFunction(parameterValue)
}

func MD5HashParameter(name string) ParameterDescription {
	return MatchingParamter(name, "[a-z0-9]{32}")
}

func UUIDParameter(name string) ParameterDescription {
	return MatchingParamter(name, "[^a-zA-Z0-9]+")
}

func MatchingParamter(name string, pattern string) ParameterDescription {
	return &parameterDescrition{
		name: name,
		validatorFunction: func(value string) error {
			if match, err := regexp.Match(pattern, []byte(value)); !match {
				return fmt.Errorf("pattern not matched")
			} else if err != nil {
				return err
			} else {
				return nil
			}
		},
	}
}

func ReadIntParameter(r *http.Request, parameterName string) (bool, int) {
	parameterMap := r.Context().Value(ContextParamsKey).(ParameterMap)
	key := ParamsKey(parameterName)
	if stringValue, ok := parameterMap[key]; !ok {
		return false, 0
	} else if intValue, err := strconv.Atoi(stringValue); err != nil {
		return false, 0
	} else {
		return true, intValue
	}
}

func ReadStringParameter(r *http.Request, parameterDescrition ParameterDescription) (bool, string) {
	ctx := r.Context()
	parameterMap := ctx.Value(ContextParamsKey).(ParameterMap)
	key := ParamsKey(parameterDescrition.getParameterName())
	if val, ok := parameterMap[key]; !ok {
		return false, val
	} else if err := parameterDescrition.validate(val); err != nil {
		return false, val
	} else {
		return true, val
	}
}

func JoinStringParameters(r *http.Request, params ParameterMap) context.Context {
	ctx := r.Context()
	oldParams := ctx.Value(ContextParamsKey)

	if oldParams == nil {
		oldParams = make(ParameterMap)
	}

	for k, v := range params {
		oldParams.(ParameterMap)[k] = v
	}

	return context.WithValue(ctx, ContextParamsKey, oldParams.(ParameterMap))
}

func shiftPart(path string) (bool, string, string) {
	if len(path) == 0 {
		return true, path, path
	} else if index := strings.Index(path, "/"); index >= 0 {
		return true, path[0:index], path[index+1:]
	} else {
		return true, path, ""
	}
}

func parseVariable(part string) (bool, string) {

	if len(part) > 1 && part[:1] == "{" && part[len(part)-1:] == "}" {
		return true, part[1 : len(part)-1]
	} else {
		return false, ""
	}
}

func MatchPath(pattern string, path string) (bool, string, ParameterMap) {

	params := make(ParameterMap)

	remaining := path

	matchedParts := make([]string, 0)

	for _, part := range strings.Split(pattern, "/") {

		if part == "**" {
			return true, path, params
		}

		if success, token, r := shiftPart(remaining); !success {
			return false, "", nil
		} else {
			remaining = r
			if isVariable, variableName := parseVariable(part); isVariable {

				if len(token) > 0 {
					parmerterKey := ParamsKey(variableName)
					params[parmerterKey] = token
				} else {
					return false, "", params
				}
			} else if token != part {
				return false, "", nil
			}
			matchedParts = append(matchedParts, token)
		}
	}
	return true, strings.Join(matchedParts, "/"), params
}
