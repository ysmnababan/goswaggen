package cmd

type IParser interface {
	GetAllHandlers() map[string]*[]string
}
