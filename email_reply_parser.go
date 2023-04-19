package email_reply_parser //nolint:stylecheck,golint

import (
	"regexp"
	"strings"
	"unicode"
)

type Line struct {
	Index           int
	Content         string
	ContentStripped string
	IsQuoted        bool
	IsEmpty         bool
}

const (
	space = ` `
	enter = "\n"
)

var dot = "."

func Parse(plainMail string) string {
	lines := plainMailToLines(plainMail)

	var plainLines []string
	if isQuoteOnTop(plainMail) {
		plainLines = getPlainLinesWithQuotedReplyOnTop(lines)
	} else {
		plainLines = getPlainLinesWithQuotedReplyOnBottom(lines)
	}

	return removeWhiteSpaceBeforeAndAfter(
		strings.Join(plainLines, enter),
	)
}

func plainMailToLines(plainMail string) []*Line {
	baseLines := strings.Split(plainMail, enter)

	// first save lines with some information we will use later on while parsing
	lines := make([]*Line, len(baseLines))
	for i, baseLine := range baseLines {
		contentStripped := removeWhitespace(baseLine)
		lines[i] = &Line{
			Index:           i,
			Content:         baseLine,
			ContentStripped: removeMarkdown(contentStripped),
			IsEmpty:         isWhitespace(contentStripped),
			IsQuoted:        strings.HasPrefix(baseLine, ">"),
		}
	}
	return lines
}

func getPlainLinesWithQuotedReplyOnBottom(lines []*Line) []string {
	//nolint:prealloc
	var finalLines []string
	for i, line := range lines {
		if isSignatureStart(i, line, lines) {
			break
		}
		multilineQuoteReply, _ := detectQuotedEmailStart(i, line, lines)
		if multilineQuoteReply {
			break
		}
		finalLines = append(finalLines, line.Content)
	}
	return finalLines
}

func getPlainLinesWithQuotedReplyOnTop(lines []*Line) []string {
	//nolint:prealloc
	var finalLines []string
	var quotedStartSeen bool
	var normalLineSeen bool
	var skipNextLine bool
	for i, line := range lines {
		multiLine, singleLine := detectQuotedEmailStart(i, line, lines)
		// start of quoted text can be ignored
		if multiLine {
			quotedStartSeen = true
			if !singleLine {
				skipNextLine = true
			}
			continue
		}

		// skip this line because this is still the quote start
		if skipNextLine {
			skipNextLine = false
			continue
		}

		if quotedStartSeen &&
			!line.IsQuoted &&
			!line.IsEmpty {
			normalLineSeen = true
		}

		if normalLineSeen && quotedStartSeen {
			if isSignatureStart(i, line, lines) {
				break
			}
			finalLines = append(finalLines, line.Content)
		}

	}
	return finalLines
}

func isQuoteOnTop(plainMail string) bool {
	plainMailStrippedWhiteSpace := removeWhitespace(plainMail)
	lines := plainMailToLines(plainMailStrippedWhiteSpace)
	for i, line := range lines {
		isQuoteStarted, _ := detectQuotedEmailStart(i, line, lines)
		if i == 0 && isQuoteStarted {
			return true
		}
	}

	return false
}

func lineBeforeAndAfter(lineIndex int, lines []*Line) (*Line, *Line) {
	var before *Line
	var after *Line
	if lineIndex > 0 {
		before = lines[lineIndex-1]
	}
	isLast := lineIndex != len(lines)-1
	if isLast {
		after = lines[lineIndex+1]
	}
	return before, after
}

func isSignatureStart(lineIndex int, line *Line, lines []*Line) bool {
	// first line is probably not the signature
	if lineIndex == 0 {
		return false
	}

	lowerLine := strings.ToLower(line.ContentStripped)

	// --
	// my name
	if isValidSignatureFormat(lowerLine) {
		return true
	}

	// e.g. with best regards,
	if detectGreetings(strings.ToLower(line.ContentStripped)) {
		return true
	}

	// smart system to detect signature + disclaimers like
	// Karen The Green
	// Graphic Designer
	// Office
	// Tel: +44423423423423
	// Fax: +44234234234234
	// karen@webby.com
	// Street 2, City, Zeeland, 4694EG, NL
	// www.thing.com
	//
	// Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged.

	if detectSignature(lineIndex, line, lines) {
		return true
	}

	// Sent from .... iphone/blackberry/galaxy etc
	if isSentFrom(lowerLine) {
		return true
	}

	return false
}

func detectQuotedEmailStart(lineIndex int, line *Line, lines []*Line) (bool, bool) {
	// Detect by quoted reply headers
	// sometimes there are line breaks within the quoted reply header
	_, after := lineBeforeAndAfter(lineIndex, lines)
	lineWithBreaksInOneLine := strings.ToLower(removeEnters(joinLineContents("", line, after)))

	multi := isQuotedEmailStart(lineWithBreaksInOneLine)
	single := isQuotedEmailStart(strings.ToLower(line.ContentStripped)) && after != nil && containsQuotedEmail(after.ContentStripped)
	if after != nil && containsQuotedEmail(after.ContentStripped) {
		single = false
	}
	// On .. wrote ..
	return multi, single
}

func isValidSignatureFormat(fullLine string) bool {
	return strings.TrimSpace(fullLine) == "--"
}

func detectSignature(lineIndex int, line *Line, lines []*Line) bool {
	// signatures mostly contains of numbers and short kind of labels with numbers after it
	// so we try to detect these kind of lines
	possible := isPossibleSignatureLine(line.ContentStripped)

	if possible {
		var matches int
		var lastMatchLineIndex int
		linesTillQuotedText := getLinesTillQuotedText(lineIndex, lines)
		for i, signatureLine := range linesTillQuotedText {
			if isPossibleSignatureLine(signatureLine.ContentStripped) {
				lastMatchLineIndex = i
				matches++
			}
		}

		// disclaimer
		possibleDisclaimer := getLinesTillQuotedText(lineIndex+lastMatchLineIndex+1, lines)
		filledDisclaimerLines := countLinesFilled(possibleDisclaimer)
		isDisclaimer := filledDisclaimerLines < 6

		filledLines := countLinesFilled(linesTillQuotedText)
		if isDisclaimer {
			filledLines -= filledDisclaimerLines
		}

		percentMatched := (float64(matches) * 100) / float64(filledLines)
		return percentMatched > 70
	}

	return false
}

func countLinesFilled(lines []*Line) int {
	var count int
	for _, line := range lines {
		if !line.IsEmpty {
			count++
		}
	}
	return count
}

func detectGreetings(line string) bool {
	// greetings but not
	if startWithOneOfButDoesNotContainMuchAfter(line, greetings, 2) {
		return true
	}

	// without first word e.g. best regards or
	// -->met<-- vriendelijke groeten
	if startWithOneOfButDoesNotContainMuchAfter(removeFirstWord(line), greetings, 2) {
		return true
	}
	return false
}

func removeFirstWord(sentence string) string {
	split := strings.Split(sentence, space)
	var a []string
	for i, w := range split {
		if i > 0 {
			a = append(a, w)
		}
	}
	return strings.Join(a, space)
}

func getLinesTillQuotedText(lineIndex int, lines []*Line) []*Line {
	var a []*Line

	for i, line := range lines {
		if i > lineIndex {
			multilineQuoteEmailStart, _ := detectQuotedEmailStart(i, line, lines)
			if multilineQuoteEmailStart {
				break
			}
			a = append(a, line)
		}
	}
	return a
}

func isPossibleSignatureLine(sentence string) bool {
	if isName(sentence) {
		return true
	}

	withoutStripe := strings.TrimPrefix(sentence, "-")
	if isName(strings.TrimSpace(withoutStripe)) {
		return true
	}

	noSpaceBetweenNumbers := removeSpacesBetweenNumbers(sentence)

	if isLogo(sentence) {
		return true
	}
	if areStripes(sentence) {
		return true
	}
	if isLabelWithValue(noSpaceBetweenNumbers) {
		return true
	}
	if isNumberSignature(noSpaceBetweenNumbers) {
		return true
	}
	if isEmailSignature(sentence) {
		return true
	}
	if isWebsiteSignature(sentence) {
		return true
	}
	return false
}

func isLogo(sentence string) bool {
	return strings.HasPrefix(sentence, "[image") && strings.HasSuffix(sentence, "]")
}

func removeSpacesBetweenNumbers(sentence string) string {
	var newSentence string
	split := strings.Split(sentence, space)
	for i, word := range split {
		if i > 0 {
			prevWord := split[i-1]
			if !isNumberWord(prevWord) {
				newSentence += space
			}
		}
		newSentence += word
	}
	return newSentence
}

func isNumberWord(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func isWebsiteSignature(sentence string) bool {
	spaces := strings.Count(sentence, space)
	return containsWebsite(sentence) && spaces <= 2
}

func isEmailSignature(sentence string) bool {
	spaces := strings.Count(sentence, space)
	return containsEmail(sentence) && spaces <= 2
}

func isNumberSignature(sentence string) bool {
	spaces := strings.Count(sentence, space)
	return containsNumber(sentence) && spaces <= 3
}

func areStripes(sentence string) bool {
	for _, c := range sentence {
		if string(c) != `-` && string(c) != `_` {
			return false
		}
	}

	return len(sentence) > 0
}

func isLabelWithValue(v string) bool {
	// is a telephone number with label or some other stuff
	lowerLine := strings.ToLower(v)
	withoutLabel := removeFirstWord(lowerLine)

	amountOfSpaces := strings.Count(withoutLabel, space)

	// Beatrixlaan 2, 4694EG Scherpenisse
	// if amountOfCommas >

	return amountOfSpaces <= 1 && (containsEmail(withoutLabel) ||
		containsWebsite(withoutLabel) ||
		isNumberSignature(withoutLabel))
}

func containsWebsite(v string) bool {
	words := strings.Split(v, space)
	for _, word := range words {
		containsSlashes := strings.Count(word, "/")
		containsDots := strings.Count(word, ".")
		containsHttp := strings.Count(word, "http")
		containsWww := strings.Count(word, "www")

		containsExtension := hasOneOf(word, extensions, &dot, nil)

		count := containsSlashes + containsDots + containsHttp + containsWww

		if count >= 1 && containsExtension {
			return true
		}
	}
	return false
}

func containsEmail(v string) bool {
	words := strings.Split(v, space)
	for _, word := range words {
		if strings.Contains(word, "@") &&
			strings.Contains(word, ".") {
			return true
		}
	}
	return false
}

func containsNumber(v string) bool {
	words := strings.Split(v, space)
	for _, word := range words {
		if amountOfDigits(word) > 5 {
			return true
		}
	}
	return false
}

func isName(sentence string) bool {
	nameAndFunction := splitNameAndFunction(sentence)

	splitName := strings.Split(removeWhitespace(nameAndFunction[0]), space)

	// is a name e.g Kate Green, Richard Lindhout, Jan van der Doorn
	if len(splitName) > 0 && len(splitName) <= 3 {
		firstName := removeWhitespace(splitName[0])
		lastName := removeWhitespace(splitName[len(splitName)-1])

		isValidName := (len(firstName) > 3 || len(lastName) > 3) && isFirstLetterUppercase(firstName) && isFirstLetterUppercase(lastName)
		invalidCharacters := containsSpecialCharacters(firstName) || containsSpecialCharacters(lastName)

		if len(nameAndFunction) > 1 {
			// function name should not be more than 3 words
			function := removeWhitespace(nameAndFunction[1])
			isFunction := len(function) > 3
			validFunction := strings.Count(function, space) <= 3

			if isFunction && !validFunction {
				return false
			}
		}

		return isValidName && !invalidCharacters
	}
	return false
}

var separators = []string{"|", "-", ","}

func splitNameAndFunction(v string) []string {
	for _, separator := range separators {
		if strings.Count(v, separator) == 1 {
			return strings.Split(v, separator)
		}
	}
	return []string{v}
}

var specialCharacters = []rune("[!@#$%&*()_+=|<>?{}[]~-]")

func containsSpecialCharacters(v string) bool {
	for _, specialC := range specialCharacters {
		for _, c := range v {
			if c == specialC {
				return true
			}
		}
	}
	return false
}

func amountOfDigits(v string) int {
	var amount int
	for _, c := range v {
		if unicode.IsDigit(c) {
			amount++
		}
	}
	return amount
}

func removeMarkdown(s string) string {
	newS := strings.Replace(s, "*", "", -1)
	return newS
}

func isFirstLetterUppercase(v string) bool {
	if len(v) > 0 {
		for i, c := range v {
			if i == 0 {
				return unicode.IsUpper(c)
			}
		}
	}
	return false
}

func isSentFrom(fullLine string) bool {
	startsWithSend := startWithOneOf(fullLine, sent, true)
	containsDevice := hasOneOf(fullLine, mailPrograms, nil, nil)
	return startsWithSend && containsDevice
}

var spaceStr = " "

func isQuotedEmailStart(fullLine string) bool {
	// on ... wrote etc
	// On Monday, November 4, 2013 4:29 PM, John Smith <john.smith@example.org> wrote:
	// Op za 8 mei 2021 om 12:09 schreef Richard Lindhout <richardlindhout96@gmail.com>:
	// On Oct 1, 2012, at 11:55 PM, Dave Tapley wrote:
	// 2013/11/1 John Smith <john@smith.org>
	startsWithOn := startWithOneOf(fullLine, on, true)
	containsWrote := hasOneOf(fullLine, wrote, &spaceStr, nil)
	allNumbers := findNumbers(fullLine)
	containsYear := numberArrayContainsYear(allNumbers)
	containsEnoughNumbers := len(allNumbers) >= 3
	// TODO: more strict
	containsQuotedEmail := strings.Contains(fullLine, "@") &&
		strings.Contains(fullLine, "<") &&
		strings.Contains(fullLine, ">")

	if startsWithOn && containsWrote && containsEnoughNumbers && containsYear {
		return true
	} else if containsQuotedEmail && containsEnoughNumbers && containsYear {
		return true
	}
	return false
}

func containsQuotedEmail(v string) bool {
	return strings.Contains(v, "@") &&
		strings.Contains(v, "<") &&
		strings.Contains(v, ">")
}

func joinLineContents(sep string, lines ...*Line) string {
	var a []string
	for _, line := range lines {
		if line != nil {
			a = append(a, line.ContentStripped)
		}
	}
	return strings.Join(a, sep)
}

//nolint:gochecknoglobals
var numberRegex = regexp.MustCompile("[0-9]+")

func findNumbers(v string) []string {
	return numberRegex.FindAllString(v, -1)
}

func numberArrayContainsYear(a []string) bool {
	for _, v := range a {
		if len(v) == 4 {
			return true
		}
	}
	return false
}

func hasOneOf(value string, a []string, addFront *string, addBack *string) bool {
	for _, c := range a {
		finalContains := strings.ToLower(c)
		if addFront != nil {
			finalContains = *addFront + finalContains
		}
		if addBack != nil {
			finalContains += *addBack
		}
		if strings.Contains(value, finalContains) {
			return true
		}
	}
	return false
}

func startWithOneOf(value string, a []string, addSpaceAfter bool) bool {
	for _, prefix := range a {
		finalPrefix := strings.ToLower(prefix)
		if addSpaceAfter {
			finalPrefix += space
		}
		if strings.HasPrefix(value, finalPrefix) {
			return true
		}
	}
	return false
}

func startWithOneOfButDoesNotContainMuchAfter(value string, a []string, maxCharactersAfter int) bool {
	for _, prefix := range a {
		finalPrefix := strings.ToLower(prefix)

		if strings.HasPrefix(value, finalPrefix) {
			after := strings.TrimPrefix(value, finalPrefix)
			return len(after) <= maxCharactersAfter
		}
	}
	return false
}

// isWhitespace returns true if the string consist of white space
func isWhitespace(content string) bool {
	// If the node is a space it's an enter
	for _, v := range content {
		// Test each character to see if it is whitespace.
		if !unicode.IsSpace(v) {
			return false
		}
	}
	return true
}

// removeWhitespace removes all double spaces in text
func removeWhitespace(v string) (r string) {
	r = strings.ReplaceAll(v, "\t", space)
	r = strings.ReplaceAll(r, `	`, space)
	r = strings.ReplaceAll(r, "  ", space)
	r = strings.ReplaceAll(r, `  `, space)
	if v != r {
		r = removeWhitespace(r)
	}
	return strings.TrimSpace(r)
}

func removeEnters(v string) (r string) {
	r = strings.ReplaceAll(v, enter, "")
	if v != r {
		r = removeEnters(r)
	}
	return strings.TrimSpace(r)
}

func removeWhiteSpaceBeforeAndAfter(v string) (r string) {
	r = strings.TrimSpace(v)
	r = strings.TrimSuffix(r, enter)
	r = strings.TrimPrefix(r, enter)
	r = strings.TrimSpace(r)

	if v != r {
		r = removeWhiteSpaceBeforeAndAfter(r)
	}
	return r
}
