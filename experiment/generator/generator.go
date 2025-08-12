package generator

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/ysmnababan/goswaggen/experiment/generator/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type generator struct {
	funcName      string
	method        string
	frameworkName string
	payloads      []*model.PayloadInfo
	responses     []*model.ReturnResponse
	commentBlock  *model.CommentBlock
}

func NewGenerator(p Parser) *generator {
	return &generator{
		funcName:      p.GetFuncName(),
		method:        strings.ToUpper(p.GetMethod()),
		frameworkName: p.GetFrameworkName(),
		payloads:      p.GetPayloadInfos(),
		responses:     p.ReturnResponses(),
		commentBlock: &model.CommentBlock{
			Params:   []string{},
			Produce:  []string{},
			Response: []string{},
		},
	}
}

func (g *generator) CreateCommentBlock() *model.CommentBlock {
	g.setSummary()
	g.setDescription()
	g.setTags()
	g.setAccept()
	g.setParam()
	g.setResponse()
	return g.commentBlock
}

func (g *generator) setSummary() {
	g.commentBlock.Summary = fmt.Sprintf("// @Summary  %s", g.funcName)
}

func camelCaseToTitle(input string) string {
	// Regex to split camel case (e.g., "CreateCommentBlock" → "Create Comment Block")
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	withSpaces := re.ReplaceAllString(input, `$1 $2`)

	// Split into words
	words := strings.Split(withSpaces, " ")
	if len(words) == 0 {
		return ""
	}

	// Use a caser to title-case the first word
	caser := cases.Title(language.English)

	words[0] = caser.String(words[0])
	for i := 1; i < len(words); i++ {
		words[i] = strings.ToLower(words[i])
	}

	return strings.Join(words, " ")
}

func (g *generator) setDescription() {
	g.commentBlock.Description = fmt.Sprintf("// @Description %s", camelCaseToTitle(g.funcName))
}

func (g *generator) setTags() {
	g.commentBlock.Tags = "// @Tags ______ "
}

func (g *generator) setAccept() {
	if g.method == "GET" || g.method == "DELETE" {
		return
	}
	acceptMap := make(map[string]bool)
	for _, p := range g.payloads {
		if tag := p.GetAcceptTag(); tag != "" {
			acceptMap[tag] = true
		}
	}
	tags := []string{}
	for k := range acceptMap {
		tags = append(tags, k)
	}
	if len(tags) > 0 {
		g.commentBlock.Accept = fmt.Sprintf("// @Accept %s", strings.Join(tags, ","))
	}
}

func (g *generator) setParam() {
	existingParam := make(map[string]bool)
	params := []*model.Param{}
	for _, p := range g.payloads {
		results := processPayload(p, g.method)
		if len(results) == 0 {
			continue
		}
		params = append(params, results...)
	}
	for _, p := range params {
		comment := fmt.Sprintf("%s %s %s %v %s", p.Name, p.BindMethod, p.ParamType, p.IsRequired, p.Description)
		if _, ok := existingParam[comment]; !ok {
			existingParam[comment] = true
			g.commentBlock.Params = append(g.commentBlock.Params, fmt.Sprintf("// @Param %s", comment))
		}
	}
}

func processPayload(i *model.PayloadInfo, method string) []*model.Param {
	out := []*model.Param{}
	switch i.BindMethod {
	case "Bind":
		if method == "GET" || method == "DELETE" {
			for _, f := range i.FieldLists {
				method, name := getPriorityTag(f.Tag)
				isRequired := true
				if !isRequiredFieldFromTag(f.Tag) && f.IsPointer {
					isRequired = false
				}
				if method != "" && name != "" {
					p := &model.Param{
						Name:        name,
						BindMethod:  method,
						ParamType:   f.VarType,
						Description: DEFAULT_PARAM_DESCRIPTION,
						IsRequired:  isRequired,
					}
					out = append(out, p)
				}
			}
		} else {
			p := &model.Param{
				Name:        i.BasicLit,
				BindMethod:  "body",
				ParamType:   i.ParamTypes,
				IsRequired:  true,
				Description: DEFAULT_BODY_DESCRIPTION,
			}
			out = append(out, p)
		}
	case "QueryParam":
		p := &model.Param{
			Name:        i.BasicLit,
			BindMethod:  "query",
			IsRequired:  true,
			ParamType:   "string",
			Description: DEFAULT_QUERY_PARAM_DESCRIPTION,
		}
		out = append(out, p)
	case "Param":
		p := &model.Param{
			Name:        i.BasicLit,
			BindMethod:  "path",
			IsRequired:  true,
			ParamType:   "string",
			Description: DEFAULT_PARAM_DESCRIPTION,
		}
		out = append(out, p)
	default:
		log.Println("unknown bind method:", i.BindMethod)
	}
	return out
}

func getPriorityTag(tags map[string]string) (method string, name string) {
	method = ""
	name = ""
	if n, ok := tags["param"]; ok {
		name = n
		method = "path"
	}
	if n, ok := tags["query"]; ok {
		name = n
		method = "query"
	}
	return
}

func isRequiredFieldFromTag(tags map[string]string) bool {
	isvalid := tags["validate"]
	return isvalid == "required"
}

func processResponse(r *model.ReturnResponse) string {
	prefix := "Failure"
	desc := DEFAULT_FAILURE_RESPONSE_DESCRIPTION
	if r.IsSuccess {
		prefix = "Success"
		desc = DEFAULT_SUCCESS_RESPONSE_DESCRIPTION
	}
	schemeType, ok := GO_TO_SWAGGO_SCHEME_TYPES_MAP[strings.ToLower(r.AcceptType)]
	if !ok {
		schemeType = DEFAULT_RESPONSE_SCHEME_TYPE
	}

	out := fmt.Sprintf(RESPONSE_BLOCK_TEMPLATE,
		prefix,
		r.StatusCode,
		schemeType,
		r.StructType,
		desc,
	)
	return out
}

func (g *generator) setResponse() {
	existingResp := make(map[string]bool)

	// the response block at least has these response,
	// if not exist, add default resp
	defaultResp := map[string]bool{
		"@Success 200": false,
		"@Failure 400": false,
		"@Failure 404": false,
		"@Failure 500": false,
	}

	for _, r := range g.responses {
		result := processResponse(r)
		_, ok := existingResp[result]
		if ok {
			// handle duplicate
			continue
		}
		for k := range defaultResp {
			if strings.Contains(result, k) {
				defaultResp[k] = true
			}
		}
		g.commentBlock.Response = append(g.commentBlock.Response, result)
	}

	// TODO: Update default value by config
	defaultSuccess := " {object} default.Success"
	defaultFailure := " {object} default.Failure"

	// add default resp if not exist
	for k, val := range defaultResp {
		if val {
			continue
		}
		if strings.Contains(k, "200") {
			g.commentBlock.Response = append(g.commentBlock.Response, k+defaultSuccess)
		} else {
			g.commentBlock.Response = append(g.commentBlock.Response, k+defaultFailure)
		}
	}
}

func (g *generator) setAcceptType() {
	isExist := make(map[string]bool)
	accepts := []string{}

	for _, r := range g.responses {
		if ok := isExist[r.AcceptType]; !ok {
			isExist[r.AcceptType] = true
			accepts = append(accepts, r.AcceptType)
		}
	}
	if len(accepts) != 0 {
		g.commentBlock.Accept = strings.Join(accepts, ",")
	}
}
