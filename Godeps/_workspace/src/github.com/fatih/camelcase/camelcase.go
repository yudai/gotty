// Package camelcase is a micro package to split the words of a camelcase type
// string into a slice of words.
package camelcase

import "unicode"

// Split splits the camelcase word and returns a list of words. It also
// supports digits.  Both lower camel case and upper camel case are supported.
// For more info please check:  http://en.wikipedia.org/wiki/CamelCase
//
// Below are some example cases:
//   lowercase =>       ["lowercase"]
//   Class =>           ["Class"]
//   MyClass =>         ["My", "Class"]
//   MyC =>             ["My", "C"]
//   HTML =>            ["HTML"]
//   PDFLoader =>       ["PDF", "Loader"]
//   AString =>         ["A", "String"]
//   SimpleXMLParser => ["Simple", "XML", "Parser"]
//   vimRPCPlugin =>    ["vim", "RPC", "Plugin"]
//   GL11Version =>     ["GL", "11", "Version"]
//   99Bottles =>       ["99", "Bottles"]
//   May5 =>            ["May", "5"]
//   BFG9000 =>         ["BFG", "9000"]
func Split(src string) []string {
	if src == "" {
		return []string{}
	}

	splitIndex := []int{}
	for i, r := range src {
		// we don't care about first index
		if i == 0 {
			continue
		}

		// search till we find an upper case
		if unicode.IsLower(r) {
			continue
		}

		prevRune := rune(src[i-1])

		// for cases like: GL11Version, BFG9000
		if unicode.IsDigit(r) && !unicode.IsDigit(prevRune) {
			splitIndex = append(splitIndex, i)
			continue
		}

		if !unicode.IsDigit(r) && !unicode.IsUpper(prevRune) {
			// for cases like: MyC
			if i+1 == len(src) {
				splitIndex = append(splitIndex, i)
				continue
			}

			// for cases like: SimpleXMLParser, eclipseRCPExt
			if unicode.IsUpper(rune(src[i+1])) {
				splitIndex = append(splitIndex, i)
				continue
			}
		}

		// If the next char is lower case, we have found a split index
		if i+1 != len(src) && unicode.IsLower(rune(src[i+1])) {
			splitIndex = append(splitIndex, i)
		}
	}

	// nothing to split, such as "hello", "Class", "HTML"
	if len(splitIndex) == 0 {
		return []string{src}
	}

	// now split the input string into pieces
	splitted := make([]string, len(splitIndex)+1)
	for i := 0; i < len(splitIndex)+1; i++ {
		if i == 0 {
			// first index
			splitted[i] = src[:splitIndex[0]]
		} else if i == len(splitIndex) {
			// last index
			splitted[i] = src[splitIndex[i-1]:]
		} else {
			// between first and last index
			splitted[i] = src[splitIndex[i-1]:splitIndex[i]]
		}
	}

	return splitted
}
