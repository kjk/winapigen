package xmldef

// generated with: find API -name "*.xml" | xargs chidley -X -a "" -e "" -t -F
// go get -u github.com/gnewton/chidley

import "encoding/xml"

type ApiMonitor struct {
	XMLName            xml.Name            `xml:"ApiMonitor,omitempty" json:"ApiMonitor,omitempty"`
	ApiSetSchema       *ApiSetSchema       `xml:"ApiSetSchema,omitempty" json:"ApiSetSchema,omitempty"`
	ErrorLookupModule  *ErrorLookupModule  `xml:"ErrorLookupModule,omitempty" json:"ErrorLookupModule,omitempty"`
	Headers            *Headers            `xml:"Headers,omitempty" json:"Headers,omitempty"`
	HelpUrl            []*HelpUrl          `xml:"HelpUrl,omitempty" json:"HelpUrl,omitempty"`
	Include            []*Include          `xml:"Include,omitempty" json:"Include,omitempty"`
	Interface          *Interface          `xml:"Interface,omitempty" json:"Interface,omitempty"`
	Module             []*Module           `xml:"Module,omitempty" json:"Module,omitempty"`
	UnsupportedModules *UnsupportedModules `xml:"UnsupportedModules,omitempty" json:"UnsupportedModules,omitempty"`
}

type Include struct {
	XMLName  xml.Name `xml:"Include,omitempty" json:"Include,omitempty"`
	Filename string   `xml:"Filename,attr"  json:",omitempty"`
}

type Module struct {
	XMLName            xml.Name        `xml:"Module,omitempty" json:"Module,omitempty"`
	CallingConvention  string          `xml:"CallingConvention,attr"  json:",omitempty"`
	Category           string          `xml:"Category,attr"  json:",omitempty"`
	ErrorFunc          string          `xml:"ErrorFunc,attr"  json:",omitempty"`
	ErrorIsReturnValue string          `xml:"ErrorIsReturnValue,attr"  json:",omitempty"`
	Name               string          `xml:"Name,attr"  json:",omitempty"`
	OnlineHelp         string          `xml:"OnlineHelp,attr"  json:",omitempty"`
	Api                []*Api          `xml:"Api,omitempty" json:"Api,omitempty"`
	Categories         []*Category     `xml:"Category,omitempty" json:"Category,omitempty"`
	Condition          []*Condition    `xml:"Condition,omitempty" json:"Condition,omitempty"`
	ErrorDecode        []*ErrorDecode  `xml:"ErrorDecode,omitempty" json:"ErrorDecode,omitempty"`
	ModuleAlias        []*ModuleAlias  `xml:"ModuleAlias,omitempty" json:"ModuleAlias,omitempty"`
	SourceModule       []*SourceModule `xml:"SourceModule,omitempty" json:"SourceModule,omitempty"`
	Variable           []*Variable     `xml:"Variable,omitempty" json:"Variable,omitempty"`
	Str                string          `xml:",chardata" json:",omitempty"`
}

type Category struct {
	XMLName    xml.Name `xml:"Category,omitempty" json:"Category,omitempty"`
	Name       string   `xml:"Name,attr"  json:",omitempty"`
	OnlineHelp string   `xml:"OnlineHelp,attr"  json:",omitempty"`
}

type Api struct {
	XMLName            xml.Name `xml:"Api,omitempty" json:"Api,omitempty"`
	BothCharset        string   `xml:"BothCharset,attr"  json:",omitempty"`
	Disabled_Discard   string   `xml:"Disabled_Discard,attr"  json:",omitempty"`
	Discard            string   `xml:"Discard,attr"  json:",omitempty"`
	Display            string   `xml:"Display,attr"  json:",omitempty"`
	ErrorFunc          string   `xml:"ErrorFunc,attr"  json:",omitempty"`
	ErrorIsReturnValue string   `xml:"ErrorIsReturnValue,attr"  json:",omitempty"`
	MaxVarArgs         string   `xml:"MaxVarArgs,attr"  json:",omitempty"`
	Name               string   `xml:"Name,attr"  json:",omitempty"`
	Ordinal            string   `xml:"Ordinal,attr"  json:",omitempty"`
	OrdinalA           string   `xml:"OrdinalA,attr"  json:",omitempty"`
	OrdinalW           string   `xml:"OrdinalW,attr"  json:",omitempty"`
	SuffixA            *string  `xml:"SuffixA,attr"  json:",omitempty"`
	VarArgs            string   `xml:"VarArgs,attr"  json:",omitempty"`
	Param              []*Param `xml:"Param,omitempty" json:"Param,omitempty"`
	Return             *Return  `xml:"Return,omitempty" json:"Return,omitempty"`
	Success            *Success `xml:"Success,omitempty" json:"Success,omitempty"`
	Str                string   `xml:",chardata" json:",omitempty"`
}

type Param struct {
	XMLName              xml.Name `xml:"Param,omitempty" json:"Param,omitempty"`
	Count                string   `xml:"Count,attr"  json:",omitempty"`
	DISABLED_InterfaceId string   `xml:"DISABLED_InterfaceId,attr"  json:",omitempty"`
	DerefCount           string   `xml:"DerefCount,attr"  json:",omitempty"`
	DerefPostCount       string   `xml:"DerefPostCount,attr"  json:",omitempty"`
	DerefPostLength      string   `xml:"DerefPostLength,attr"  json:",omitempty"`
	Display              string   `xml:"Display,attr"  json:",omitempty"`
	InterfaceId          string   `xml:"InterfaceId,attr"  json:",omitempty"`
	Length               string   `xml:"Length,attr"  json:",omitempty"`
	Name                 string   `xml:"Name,attr"  json:",omitempty"`
	OutputOnly           string   `xml:"OutputOnly,attr"  json:",omitempty"`
	PostCount            string   `xml:"PostCount,attr"  json:",omitempty"`
	PostLength           string   `xml:"PostLength,attr"  json:",omitempty"`
	Type                 string   `xml:"Type,attr"  json:",omitempty"`
}

type Return struct {
	XMLName xml.Name `xml:"Return,omitempty" json:"Return,omitempty"`
	Display string   `xml:"Display,attr"  json:",omitempty"`
	Type    string   `xml:"Type,attr"  json:",omitempty"`
}

type Variable struct {
	XMLName    xml.Name `xml:"Variable,omitempty" json:"Variable,omitempty"`
	Base       string   `xml:"Base,attr"  json:",omitempty"`
	Count      string   `xml:"Count,attr"  json:",omitempty"`
	DisplayHex string   `xml:"DisplayHex,attr"  json:",omitempty"`
	Name       string   `xml:"Name,attr"  json:",omitempty"`
	Pack       string   `xml:"Pack,attr"  json:",omitempty"`
	Pack32     string   `xml:"Pack32,attr"  json:",omitempty"`
	Pack64     string   `xml:"Pack64,attr"  json:",omitempty"`
	Size       string   `xml:"Size,attr"  json:",omitempty"`
	Type       string   `xml:"Type,attr"  json:",omitempty"`
	Unsigned   string   `xml:"Unsigned,attr"  json:",omitempty"`
	Display    *Display `xml:"Display,omitempty" json:"Display,omitempty"`
	Enum       *Enum    `xml:"Enum,omitempty" json:"Enum,omitempty"`
	Field      []*Field `xml:"Field,omitempty" json:"Field,omitempty"`
	Flag       *Flag    `xml:"Flag,omitempty" json:"Flag,omitempty"`
	Set        []*Set   `xml:"Set,omitempty" json:"Set,omitempty"`
	Success    *Success `xml:"Success,omitempty" json:"Success,omitempty"`
	Str        string   `xml:",chardata" json:",omitempty"`
}

type Display struct {
	XMLName xml.Name `xml:"Display,omitempty" json:"Display,omitempty"`
	Name    string   `xml:"Name,attr"  json:",omitempty"`
}

type Enum struct {
	XMLName     xml.Name `xml:"Enum,omitempty" json:"Enum,omitempty"`
	DefaultName string   `xml:"DefaultName,attr"  json:",omitempty"`
	Reset       string   `xml:"Reset,attr"  json:",omitempty"`
	Set         []*Set   `xml:"Set,omitempty" json:"Set,omitempty"`
}

type Set struct {
	XMLName  xml.Name `xml:"Set,omitempty" json:"Set,omitempty"`
	Name     string   `xml:"Name,attr"  json:",omitempty"`
	Value    string   `xml:"Value,attr"  json:",omitempty"`
	ValueInt int
}

type Success struct {
	XMLName   xml.Name `xml:"Success,omitempty" json:"Success,omitempty"`
	ErrorFunc string   `xml:"ErrorFunc,attr"  json:",omitempty"`
	Return    string   `xml:"Return,attr"  json:",omitempty"`
	Value     string   `xml:"Value,attr"  json:",omitempty"`
}

type HelpUrl struct {
	XMLName xml.Name `xml:"HelpUrl,omitempty" json:"HelpUrl,omitempty"`
	Display string   `xml:"Display,attr"  json:",omitempty"`
	Name    string   `xml:"Name,attr"  json:",omitempty"`
	Url     string   `xml:"Url,attr"  json:",omitempty"`
}

type Headers struct {
	XMLName   xml.Name     `xml:"Headers,omitempty" json:"Headers,omitempty"`
	Condition []*Condition `xml:"Condition,omitempty" json:"Condition,omitempty"`
	Variable  []*Variable  `xml:"Variable,omitempty" json:"Variable,omitempty"`
	Str       string       `xml:",chardata" json:",omitempty"`
}

type Field struct {
	XMLName    xml.Name `xml:"Field,omitempty" json:"Field,omitempty"`
	Count      string   `xml:"Count,attr"  json:",omitempty"`
	DerefCount string   `xml:"DerefCount,attr"  json:",omitempty"`
	Display    string   `xml:"Display,attr"  json:",omitempty"`
	Length     string   `xml:"Length,attr"  json:",omitempty"`
	Name       string   `xml:"Name,attr"  json:",omitempty"`
	OutputOnly string   `xml:"OutputOnly,attr"  json:",omitempty"`
	PostCount  string   `xml:"PostCount,attr"  json:",omitempty"`
	PostLength string   `xml:"PostLength,attr"  json:",omitempty"`
	Type       string   `xml:"Type,attr"  json:",omitempty"`
}

type Flag struct {
	XMLName xml.Name `xml:"Flag,omitempty" json:"Flag,omitempty"`
	Reset   string   `xml:"Reset,attr"  json:",omitempty"`
	Set     []*Set   `xml:"Set,omitempty" json:"Set,omitempty"`
}

type Interface struct {
	XMLName       xml.Name    `xml:"Interface,omitempty" json:"Interface,omitempty"`
	BaseInterface string      `xml:"BaseInterface,attr"  json:",omitempty"`
	Category      string      `xml:"Category,attr"  json:",omitempty"`
	ErrorFunc     string      `xml:"ErrorFunc,attr"  json:",omitempty"`
	Id            string      `xml:"Id,attr"  json:",omitempty"`
	Name          string      `xml:"Name,attr"  json:",omitempty"`
	OnlineHelp    string      `xml:"OnlineHelp,attr"  json:",omitempty"`
	Api           []*Api      `xml:"Api,omitempty" json:"Api,omitempty"`
	Variable      []*Variable `xml:"Variable,omitempty" json:"Variable,omitempty"`
}

type ModuleAlias struct {
	XMLName xml.Name `xml:"ModuleAlias,omitempty" json:"ModuleAlias,omitempty"`
	Name    string   `xml:"Name,attr"  json:",omitempty"`
}

type ErrorDecode struct {
	XMLName            xml.Name `xml:"ErrorDecode,omitempty" json:"ErrorDecode,omitempty"`
	ErrorFunc          string   `xml:"ErrorFunc,attr"  json:",omitempty"`
	ErrorIsReturnValue string   `xml:"ErrorIsReturnValue,attr"  json:",omitempty"`
}

type ErrorLookupModule struct {
	XMLName xml.Name `xml:"ErrorLookupModule,omitempty" json:"ErrorLookupModule,omitempty"`
	Name    string   `xml:"Name,attr"  json:",omitempty"`
}

type ApiSetSchema struct {
	XMLName xml.Name  `xml:"ApiSetSchema,omitempty" json:"ApiSetSchema,omitempty"`
	ApiSet  []*ApiSet `xml:"ApiSet,omitempty" json:"ApiSet,omitempty"`
}

type ApiSet struct {
	XMLName     xml.Name     `xml:"ApiSet,omitempty" json:"ApiSet,omitempty"`
	Module      string       `xml:"Module,attr"  json:",omitempty"`
	Redirection *Redirection `xml:"Redirection,omitempty" json:"Redirection,omitempty"`
}

type Redirection struct {
	XMLName xml.Name `xml:"Redirection,omitempty" json:"Redirection,omitempty"`
	Module  string   `xml:"Module,attr"  json:",omitempty"`
}

type UnsupportedModules struct {
	XMLName xml.Name  `xml:"UnsupportedModules,omitempty" json:"UnsupportedModules,omitempty"`
	Module  []*Module `xml:"Module,omitempty" json:"Module,omitempty"`
}

type Condition struct {
	XMLName      xml.Name    `xml:"Condition,omitempty" json:"Condition,omitempty"`
	Architecture string      `xml:"Architecture,attr"  json:",omitempty"`
	Variable     []*Variable `xml:"Variable,omitempty" json:"Variable,omitempty"`
}

type SourceModule struct {
	XMLName xml.Name `xml:"SourceModule,omitempty" json:"SourceModule,omitempty"`
	Copy    string   `xml:"Copy,attr"  json:",omitempty"`
	Include string   `xml:"Include,attr"  json:",omitempty"`
	Name    string   `xml:"Name,attr"  json:",omitempty"`
	Api     []*Api   `xml:"Api,omitempty" json:"Api,omitempty"`
}