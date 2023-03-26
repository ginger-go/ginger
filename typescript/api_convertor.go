package typescript

import (
	"log"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
)

func NewApiConverter() *ApiConverter {
	return &ApiConverter{
		apis: make(map[string]Api),
	}
}

type Api struct {
	Method     string
	Route      string
	Request    interface{}
	Response   interface{}
	Handler    interface{}
	Pagination bool
	Sort       bool
}

type ApiConverter struct {
	apis map[string]Api
}

func (c *ApiConverter) Add(method string, route string, request interface{}, response interface{}, handler interface{}, pagination bool, sort bool) {
	c.apis[method+":"+route] = Api{
		Method:     method,
		Route:      route,
		Request:    request,
		Response:   response,
		Handler:    handler,
		Pagination: pagination,
		Sort:       sort,
	}
}

func (c *ApiConverter) ToString() string {
	output := ""

	var names = make([]string, 0)
	for name := range c.apis {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		output += c.convertToApi(c.apis[name])
	}

	return prefix + output
}

func (c *ApiConverter) convertToApi(a Api) string {
	switch a.Method {
	case "GET":
		return c.convertToGet(a)
	case "POST":
		return c.convertToNonGet(a, "post")
	case "PUT":
		return c.convertToNonGet(a, "put")
	case "DELETE":
		return c.convertToNonGet(a, "del")
	default:
		log.Println("[WARNING] api: unknown method")
	}
	return ""
}

func (c *ApiConverter) convertToGet(a Api) string {
	if a.Request != nil {
		uriList := c.getUriList(a.Request)
		if len(uriList) > 0 {
			a.Route = c.replaceUri(a.Route, uriList)
		} else {
			a.Route += "\""
		}
	} else {
		a.Route += "\""
	}
	output := "export const " + c.nameOfFunc(a.Handler) + " = async ("

	param := "host: string"
	if a.Request != nil {
		name := c.nameOfModel(a.Request)
		if len(name) > 0 {
			output += ", req: model." + name
		}
	}
	if a.Pagination {
		param += ", page: number, size: number"
	}
	if a.Sort {
		param += ", sortBy: string, asc: boolean"
	}
	output += param + ", headers?: any): Promise<[Response<"

	if a.Response != nil {
		output += "model." + c.nameOfModel(a.Response)
	} else {
		output += "null"
	}
	output += "> | null, number]> => {\n"
	output += "    return get<"
	if a.Response != nil {
		output += "model." + c.nameOfModel(a.Response)
	} else {
		output += "null"
	}
	output += ">(host, \"" + a.Route

	if a.Request != nil || a.Pagination || a.Sort {
		var formList = make([][]string, 0)
		if a.Request != nil {
			formList = c.getQueryList(a.Request)
		}
		if len(formList) > 0 || a.Pagination || a.Sort {
			output += ", [\n"
			for _, form := range formList {
				output += "        [\"" + form[0] + "\", req." + form[1] + "],\n"
			}
			if a.Pagination {
				output += "        [\"page\", page.toString()],\n"
				output += "        [\"size\", size.toString()],\n"
			}
			if a.Sort {
				output += "        [\"sortBy\", sortBy.toString()],\n"
				output += "        [\"asc\", asc ? \"true\" : \"false\"],\n"
			}
			output += "    ]"
		}
	}
	output += ", headers)\n"
	output += "}\n\n"
	return output
}

func (c *ApiConverter) convertToNonGet(a Api, method string) string {
	if a.Request != nil {
		uriList := c.getUriList(a.Request)
		if len(uriList) > 0 {
			a.Route = c.replaceUri(a.Route, uriList)
		} else {
			a.Route += "\""
		}
	} else {
		a.Route += "\""
	}
	output := "export const " + c.nameOfFunc(a.Handler) + " = async (host: string, "
	if a.Request != nil {
		name := c.nameOfModel(a.Request)
		if len(name) > 0 {
			output += "req: model." + name + ", "
		}
	}
	output += "headers?: any): Promise<[Response<"
	if a.Response != nil {
		output += "model." + c.nameOfModel(a.Response)
	} else {
		output += "null"
	}
	output += "> | null, number]> => {\n"
	output += "    return " + method + "<"
	if a.Response != nil {
		output += "model." + c.nameOfModel(a.Response)
	} else {
		output += "null"
	}
	output += ">(host, \"" + a.Route

	if a.Request != nil {
		output += ", req, headers"
	}
	output += ")\n"
	output += "}\n\n"
	return output
}

func (c *ApiConverter) nameOfModel(model interface{}) string {
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		model = reflect.ValueOf(model).Elem().Interface()
	}
	if reflect.TypeOf(model).Kind() == reflect.Slice {
		return reflect.TypeOf(model).Elem().Name() + "[]"
	}
	return reflect.TypeOf(model).Name()
}

func (c *ApiConverter) nameOfFunc(f interface{}) string {
	xs := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), ".")
	return strings.TrimSuffix(strcase.ToLowerCamel(xs[len(xs)-1]), "Fm")
}

func (c *ApiConverter) getQueryList(model interface{}) [][]string {
	output := make([][]string, 0)
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		model = reflect.ValueOf(model).Elem().Interface()
	}
	if reflect.TypeOf(model).Kind() != reflect.Struct {
		return output
	}

	for i := 0; i < reflect.TypeOf(model).NumField(); i++ {
		tag := reflect.TypeOf(model).Field(i).Tag.Get("form")
		if tag != "" {
			output = append(output, []string{tag, tag})
		}
	}
	return output
}

func (c *ApiConverter) getUriList(model interface{}) [][]string {
	output := make([][]string, 0)
	if reflect.TypeOf(model).Kind() == reflect.Ptr {
		model = reflect.ValueOf(model).Elem().Interface()
	}
	if reflect.TypeOf(model).Kind() != reflect.Struct {
		return output
	}

	for i := 0; i < reflect.TypeOf(model).NumField(); i++ {
		tag := reflect.TypeOf(model).Field(i).Tag.Get("uri")
		if tag != "" {
			output = append(output, []string{tag, tag})
		}
	}
	return output
}

func (c *ApiConverter) replaceUri(route string, uriList [][]string) string {
	for j, uri := range uriList {
		xs := strings.Split(route, "/")

		temp := "/"
		for i, x := range xs {
			if x == "" {
				continue
			}
			if x == ":"+uri[0] {
				temp += "\" + req." + uri[1]
				if i != len(xs)-1 {
					temp += " + \"/"
				}
			} else {
				temp += x
				if i != len(xs)-1 {
					temp += "/"
				} else if temp[len(temp)-1] != '"' && j == len(uriList)-1 {
					temp += "\""
				}
			}
		}
		route = temp
	}
	return route
}

const prefix = `
import * as model from './model';

export interface Response<T> {
    success: boolean;
    error?: Error;
    pagination?: Pagination;
    data?: T;
}

export interface Pagination {
    page: number;
    size: number;
    total: number;
}

export interface Error {
    code: string;
    message: string;
}

export const get = async <T>(host: string, url: string, params?: any[][], headers?: any): Promise<[Response<T> | null, number]> => {
    try {
        url = host + url;
        if (params) {
            var li: string[] = []
            params.map(([key, value]) => {
                if (key !== undefined && key !== null && value !== undefined && value !== null) {
                    li.push(key + "=" + value)
                }
            })
            url += '?' + li.join('&')
        }
        const response = await fetch(url, {
            method: 'GET',
            headers: headers
        });
        return _handleResponse(response);
    } catch (err) {
        console.error(err);
        return [null, 0];
    }
}

export const post = async <T>(host: string, url: string, body?: any, headers?: any): Promise<[Response<T> | null, number]> => {
    return await _nonGet('POST', host + url, body, headers);
}

export const put = async <T>(host: string, url: string, body?: any, headers?: any): Promise<[Response<T> | null, number]> => {
    return await _nonGet('PUT', host + url, body, headers);
}

export const del = async <T>(host: string, url: string, body?: any, headers?: any): Promise<[Response<T> | null, number]> => {
    return await _nonGet('DELETE', host + url, body, headers);
}

export const upload = async <T>(host: string, url: string, file: File, headers?: any): Promise<[Response<T> | null, number]> => {
    try {
        const formData = new FormData();
        formData.append('file', file);
        const response = await fetch(host + url, {
            method: 'POST',
            headers: headers,
            body: formData,
        });
        return _handleResponse(response);
    } catch (err) {
        console.error(err);
        return [null, 0];
    }
}

const _nonGet = async <T>(host: string, method: string, url: string, body?: any, headers?: any): Promise<[Response<T> | null, number]> => {
    try {
        if (headers === undefined || headers === null) {
            headers = {
                "Content-Type": "application/json",
            }
        } else {
            headers["Content-Type"] = "application/json";
        }
        const response = await fetch(host + url, {
            method: method,
            headers: headers,
            body: JSON.stringify(body),
        });
        return _handleResponse(response);
    } catch (err) {
        console.error(err);
        return [null, 0];
    }
}

const _handleResponse = async <T>(resp: globalThis.Response): Promise<[T | null, number]> => {
    if (resp.status === 200) {
        return [await resp.json(), resp.status];
    }
    return [null, resp.status];
}

`
