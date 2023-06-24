package wildcard

import (
	"regexp"
	"strings"
)

var replaceMap = map[byte]string{
	'+': `\+`,
	')': `\)`,
	'$': `\$`,
	'.': `\.`,
	'{': `\{`,
	'}': `\}`,
	'|': `\|`,
	'*': ".*",
	'?': ".",
}

type Pattern struct {
	exp *regexp.Regexp
}

func (p *Pattern) IsMatch(str string) bool {
	return p.exp.MatchString(str)
}

func CompilePattern(src string) (*Pattern, error) {
	regSrc := &strings.Builder{}
	regSrc.WriteByte('^')
	for i := 0; i < len(src); i++ {
		u := src[i]
		if u == '\\' {
			regSrc.WriteByte(u)
			regSrc.WriteByte(src[i+1])
			i++
		} else if u == '^' {
			if i == 0 {
				regSrc.WriteString("\\^")
			} else {
				if src[i-1] == '[' {
					regSrc.WriteByte('^')
				} else {
					regSrc.WriteString("\\^")
				}
			}
		} else if rep, ok := replaceMap[u]; ok {
			regSrc.WriteString(rep)
		} else {
			regSrc.WriteByte(u)
		}
	}
	regSrc.WriteByte('$')
	compile, err := regexp.Compile(regSrc.String())
	if err != nil {
		return nil, err
	}
	return &Pattern{
		exp: compile,
	}, nil
}
