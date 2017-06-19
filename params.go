package goagen_js

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
)

type ServiceDefinition struct {
	Name     string
	Resource *design.ResourceDefinition
	RPCs     []RPCDefinition
}

type RPCDefinition struct {
	Action   *design.ActionDefinition
	Base     string
	Name     string
	Query    Params // sorted by alphabetical
	Response *Response
}

type Response struct {
	Name           string
	Identifier     string
	IdentifierName string
	Stream         bool // CollectionOf convert to stream
	Params         Params
}

func parseResources(g *Generator) ([]ServiceDefinition, error) {
	ret := make([]ServiceDefinition, 0)
	resources := getResources(g)

	responses, err := parseMediaTypes(g)
	if err != nil {
		return ret, err
	}

	keys := []string{}
	for n := range resources {
		keys = append(keys, n)
	}
	sort.Strings(keys)
	for _, n := range keys {
		r := resources[n]
		actions := getActions(r)

		rpcs := make([]RPCDefinition, len(actions))
		for i, action := range actions {
			rpc, err := parseAction(action, responses)
			if err != nil {
				return ret, err
			}
			rpcs[i] = rpc
		}
		ret = append(ret, ServiceDefinition{
			Resource: r,
			Name:     r.Name,
			RPCs:     rpcs,
		})
	}

	return ret, nil
}

func parseAction(action *design.ActionDefinition, responses map[string]Response) (RPCDefinition, error) {
	name := codegen.Goify(action.Name, true)

	ret := RPCDefinition{
		Action: action,
		Name:   name,
		Query:  make(Params, 0),
	}

	if action.PathParams() != nil {
		m := action.PathParams().Type.ToObject()
		for a, att := range m {
			ret.Query = append(ret.Query, newParam(action, a, att))
		}
	}

	if action.QueryParams != nil {
		m := action.QueryParams.Type.ToObject()
		for a, att := range m {
			ret.Query = append(ret.Query, newParam(action, a, att))
		}
	}

	// Payload and Query are stored same Query field.
	if action.Payload != nil {
		m := action.Payload.Type.ToObject()
		for a, att := range m {
			ret.Query = append(ret.Query, newParam(action, a, att))
		}
	}

	sort.Sort(AlphabeticalName(ret.Query))

	ret.Response = getResponse(action, responses)

	return ret, nil
}

func parseMediaTypes(g *Generator) (map[string]Response, error) {
	ret := make(map[string]Response)

	err := g.API.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.IsError() {
			return nil
		}
		err := mt.IterateViews(func(view *design.ViewDefinition) error {
			params := make(Params, 0)
			for name, att := range view.Type.ToObject() {
				p := Param{
					original:      att,
					Name:          codegen.Goify(name, false),
					CamelCaseName: codegen.Goify(name, true),
					Kind:          convertTypeString(att),
					Description:   att.Description,
				}
				params = append(params, p)
			}
			id, stream := toIdentifierName(mt.Identifier)
			ret[mt.Identifier] = Response{
				Identifier:     mt.Identifier,
				IdentifierName: id,
				Stream:         stream,
				Params:         params,
			}

			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return ret, err
}

func getResponse(action *design.ActionDefinition, responses map[string]Response) *Response {
	for _, resp := range action.Responses {
		if resp.MediaType == "" {
			continue
		}
		r, ok := responses[resp.MediaType]
		if !ok {
			continue
		}
		return &r
	}

	return nil
}

func (s ServiceDefinition) ServiceName() string {
	return codegen.Goify(s.Name, true) + "Service"
}

func (s ServiceDefinition) GetRPCs() []string {
	buf := make([]string, 0)
	for _, rpc := range s.RPCs {
		buf = append(buf, fmt.Sprintf("rpc %s(%s) returns (%s);", rpc.Name, rpc.RequestName(), rpc.ResponseName()))
	}
	return buf
}

func (p RPCDefinition) Comment() string {
	return ""
}

func (p RPCDefinition) FuncName() string {
	return p.Name
}
func (p RPCDefinition) RequestName() string {
	if len(p.Query) == 0 {
		return "Empty"
	}
	return codegen.Goify(p.Action.Parent.Name, true) + p.Name + "Type"
}
func (p RPCDefinition) ResponseName() string {
	if p.Response == nil {
		return "Empty"
	}
	if p.Response.Stream {
		return "stream " + p.Response.IdentifierName
	}

	return p.Response.IdentifierName
}

func (p RPCDefinition) RequestDefinition() []string {
	buf := make([]string, 0)
	for i, tmp := range p.Query {
		i = i + 1 // start from 1,
		buf = append(buf, tmp.MessageField(i))
	}

	return buf
}

func (p RPCDefinition) ResponseDefinition() []string {
	buf := make([]string, 0)
	if p.Response.Params == nil {
		return []string{}
	}
	for i, tmp := range p.Response.Params {
		i = i + 1 // start from 1,
		buf = append(buf, tmp.MessageField(i))
	}
	return buf
}

func repeatable(t design.Kind) bool {
	switch t {
	case design.AnyKind, design.ArrayKind:
		return true
	}
	return false
}

func convertTypeString(att *design.AttributeDefinition) string {
	if att.Metadata != nil {
		t, ok := att.Metadata["struct:field:grpctype"]
		if ok {
			return t[0]
		}
	}

	switch att.Type.Kind() {
	case design.BooleanKind:
		return "bool"
	case design.IntegerKind:
		return "int32"
	case design.NumberKind:
		return "number"
	case design.StringKind:
		return "string"
	case design.DateTimeKind:
		return "datetime"
	case design.UUIDKind:
		return "string"
	case design.AnyKind:
		return "any"
	case design.ArrayKind:
		return "array"
	case design.ObjectKind:
		return "object"
	case design.HashKind:
		return "hash"
	case design.UserTypeKind:
		return "any"
	case design.MediaTypeKind:
		return "any"
	}
	return "any"
}

type Param struct {
	original      *design.AttributeDefinition
	Name          string // no CamelCase name
	CamelCaseName string // CamelCase name
	Kind          string // kind such as bool, number, ...
	Description   string
	Enum          []string // Enum
	Repeat        bool
}

func (p Param) MessageField(i int) string {
	var buf string
	if len(p.Enum) > 0 {
		buf = p.EnumField()
	}
	buf += fmt.Sprintf("%s %s = %d;", p.Kind, p.Name, i)

	if p.Description != "" {
		buf += " // " + p.Description
	}

	return buf

}

func (p Param) EnumField() string {
	buf := fmt.Sprintf("enum %s {")
	for i, e := range p.Enum {
		// enum start from 0
		buf += fmt.Sprintf("%s = %d;", e, i)
	}
	buf += "}"

	return buf

}

type Params []Param

type AlphabeticalName []Param

func (a AlphabeticalName) Len() int           { return len(a) }
func (a AlphabeticalName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AlphabeticalName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func newParam(action *design.ActionDefinition, a string, att *design.AttributeDefinition) Param {
	p := Param{
		original:      att,
		Name:          codegen.Goify(a, false),
		CamelCaseName: codegen.Goify(a, true),
		Kind:          convertTypeString(att),
		Repeat:        repeatable(att.Type.Kind()),
		Description:   att.Description,
	}

	if att.Validation != nil && att.Validation.Values != nil {
		p.Enum = enumValues(att.Validation.Values)
	}

	return p
}

func enumValues(value interface{}) []string {
	t, ok := value.([]string)
	if !ok {
		return []string{}
	}
	return t
}

// toIdentifierName convert Identifier to CamelCase Name
// ex: application/vnd.user+json -> User
// ex: application/vnd.user+json; type=collection -> User with stream true
func toIdentifierName(identifier string) (string, bool) {
	canonical := design.CanonicalIdentifier(identifier)
	ret := strings.Replace(canonical, "; type=collection", "", 1)
	ret = strings.Replace(ret, "application/vnd.", "", 1)
	ret = codegen.Goify(ret, true)

	return ret, strings.Contains(canonical, "type=collection")
}

func getResources(g *Generator) map[string]*design.ResourceDefinition {
	ret := make(map[string]*design.ResourceDefinition)
	g.API.IterateResources(func(res *design.ResourceDefinition) error {
		ret[res.Name] = res
		return nil
	})
	return ret
}

func getActions(res *design.ResourceDefinition) []*design.ActionDefinition {
	ret := make([]*design.ActionDefinition, 0)
	res.IterateActions(func(action *design.ActionDefinition) error {
		ret = append(ret, action)
		return nil
	})
	return ret
}
