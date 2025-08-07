package generator

import (
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
		method:        p.GetMethod(),
		frameworkName: p.GetFrameworkName(),
		payloads:      p.GetPayloadInfos(),
		responses:     p.ReturnResponses(),
		commentBlock:  &model.CommentBlock{},
	}
}

func (g *generator) CreateCommentBlock() *model.CommentBlock {
	g.setSummary()
	g.setDescription()
	return g.commentBlock
}

func (g *generator) setSummary() {
	g.commentBlock.Summary = g.funcName
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
	g.commentBlock.Description = camelCaseToTitle(g.funcName)
}
