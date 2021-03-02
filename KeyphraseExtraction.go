package KeyphraseExtraction

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/jdkato/prose"
	"github.com/kljensen/snowball/english"
)

var punctuations map[string]bool
var stopWords map[string]bool
var reNumber *regexp.Regexp
var reRomanNumber *regexp.Regexp
var reHyphenedWords *regexp.Regexp
var romanHundredParts [9]string
var romanTenParts [9]string
var romanOneParts [9]string

func init() {
	punctuations = make(map[string]bool)
	punctuations["、"] = true
	punctuations[","] = true
	punctuations["，"] = true
	punctuations[":"] = true
	punctuations["："] = true
	punctuations["."] = true
	punctuations["。"] = true
	punctuations["‧"] = true
	punctuations["!"] = true
	punctuations["！"] = true
	punctuations["?"] = true
	punctuations["？"] = true
	punctuations[";"] = true
	punctuations["；"] = true
	punctuations["("] = true
	punctuations["（"] = true
	punctuations[")"] = true
	punctuations["）"] = true
	punctuations["'"] = true
	punctuations["‘"] = true
	punctuations["’"] = true
	punctuations["\""] = true
	punctuations["「"] = true
	punctuations["」"] = true
	punctuations["“"] = true
	punctuations["”"] = true
	punctuations["`"] = true
	punctuations["…"] = true

	stopWords = make(map[string]bool)
	stopWords["a"] = true
	stopWords["an"] = true
	stopWords["and"] = true
	stopWords["as"] = true
	stopWords["based"] = true
	stopWords["by"] = true
	stopWords["for"] = true
	stopWords["from"] = true
	stopWords["in"] = true
	stopWords["on"] = true
	stopWords["of"] = true
	stopWords["that"] = true
	stopWords["the"] = true
	stopWords["this"] = true
	stopWords["to"] = true
	stopWords["via"] = true
	stopWords["with"] = true
	stopWords["without"] = true

	reNumber = regexp.MustCompile("^[0-9]+$")
	reRomanNumber = regexp.MustCompile("^[iIvVxXlLcCdDmM]+$")
	reHyphenedWords = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z][a-zA-Z0-9]*)+$")

	// roman hundred parts: "cm","dccc","dcc","dc","d","cd","ccc","cc","c"
	romanHundredParts[0] = "cm"
	romanHundredParts[1] = "dccc"
	romanHundredParts[2] = "dcc"
	romanHundredParts[3] = "dc"
	romanHundredParts[4] = "d"
	romanHundredParts[5] = "cd"
	romanHundredParts[6] = "ccc"
	romanHundredParts[7] = "cc"
	romanHundredParts[8] = "c"

	// roman ten parts: "xc","lxxx","lxx","lx","l","xl","xxx","xx","x"
	romanTenParts[0] = "xc"
	romanTenParts[1] = "lxxx"
	romanTenParts[2] = "lxx"
	romanTenParts[3] = "lx"
	romanTenParts[4] = "l"
	romanTenParts[5] = "xl"
	romanTenParts[6] = "xxx"
	romanTenParts[7] = "xx"
	romanTenParts[8] = "x"

	// roman one parts: "ix","viii","vii","vi","v","iv","iii","ii","i"
	romanOneParts[0] = "ix"
	romanOneParts[1] = "viii"
	romanOneParts[2] = "vii"
	romanOneParts[3] = "vi"
	romanOneParts[4] = "v"
	romanOneParts[5] = "iv"
	romanOneParts[6] = "iii"
	romanOneParts[7] = "ii"
	romanOneParts[8] = "i"
}

/*
# =================================================================================================
# function tokenizeIntoWords
# brief description:
#   Tokenize the input text into words and puntuations, then remove the puntuations.
# input:
#   text: The input text.
# output:
#   The tokens of the text grouped by phrases separated by puntuations.
*/
func tokenizeIntoWords(text string) [][]string {
	// --------------------------------------------------------------------------------------------
	// step 1: Tokenize the input text into words and puntuations.
	tokenizer := prose.NewIterTokenizer()
	toks := tokenizer.Tokenize(text)

	// --------------------------------------------------------------------------------------------
	// step 2: Remove the puntuations and group the words into phrases separated by the puntuations.
	result := [][]string{make([]string, 0)}
	for _, tok := range toks {
		_, isPunctuation := punctuations[tok.Text]
		numPhrases := len(result)
		numWordsInLastPhrase := len(result[numPhrases-1])
		if !isPunctuation {
			if numWordsInLastPhrase > 0 {
				// recorrect some incorrect tokenization
				prevWord := result[numPhrases-1][numWordsInLastPhrase-1]
				if tok.Text == "D" || tok.Text == "d" && reNumber.MatchString(prevWord) {
					result[numPhrases-1][numWordsInLastPhrase-1] = prevWord + "d"
				} else {
					result[numPhrases-1] = append(result[numPhrases-1], tok.Text)
				}
			} else {
				result[numPhrases-1] = append(result[numPhrases-1], tok.Text)
			}
		} else if numWordsInLastPhrase > 0 {
			result = append(result, make([]string, 0))
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 2: Return the result.
	numPhrases := len(result)
	numWordsInLastPhrase := len(result[numPhrases-1])
	if numWordsInLastPhrase == 0 {
		result = result[:numPhrases-1]
	}
	return result
}

/*
# =================================================================================================
# function convertRomanToArabic
# brief description:
#   Convert a roman number to an arabic number.
# input:
#   text: The input text.
# output:
#   If the input text is a roman number, return the arabic version of this number; otherwise return
#   the original input text.
*/
func convertRomanToArabic(text string) string {
	// --------------------------------------------------------------------------------------------
	// step 1: Test the test against the regex of roman number
	looksLikeRomanNumber := reRomanNumber.MatchString(text)
	if !looksLikeRomanNumber {
		return text
	}

	// --------------------------------------------------------------------------------------------
	// step 2: Prepare the auxiliar variables.
	// numParsed: the number of chars that have been parsed.
	// numChars: the number of chars in text
	// theNumber: the output number
	numParsed := 0
	numChars := len(text)
	theNumber := 0

	// --------------------------------------------------------------------------------------------
	// step 3: convert the text to lowercase
	lowercaseText := strings.ToLower(text)

	// --------------------------------------------------------------------------------------------
	// step 4: extract the thousand part
	for numParsed < numChars {
		if lowercaseText[numParsed] != 'm' {
			break
		}
		theNumber += 1000
		numParsed++
	}

	// --------------------------------------------------------------------------------------------
	// step 5: extract the hundred part
	for idxPart, part := range romanHundredParts {
		lenPart := len(part)
		if numParsed+lenPart > numChars {
			continue
		}
		if lowercaseText[numParsed:numParsed+lenPart] == part {
			numParsed += lenPart
			theNumber += (9 - idxPart) * 100
			break
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 6: extract the ten part
	for idxPart, part := range romanTenParts {
		lenPart := len(part)
		if numParsed+lenPart > numChars {
			continue
		}
		if lowercaseText[numParsed:numParsed+lenPart] == part {
			numParsed += lenPart
			theNumber += (9 - idxPart) * 10
			break
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 7: extract the one part
	for idxPart, part := range romanOneParts {
		lenPart := len(part)
		if numParsed+lenPart > numChars {
			continue
		}
		if lowercaseText[numParsed:numParsed+lenPart] == part {
			numParsed += lenPart
			theNumber += 9 - idxPart
			break
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 8: double check: it is a roman number only if all chars have been parsed
	if numParsed == numChars {
		return fmt.Sprintf("%d", theNumber)
	} else {
		return text
	}
}

/*
# =================================================================================================
# function convertNonAbbreviationToLowercase
# brief description :
#   Convert the input text to lowercase text if it is not an abbreviation.
# input:
#   text: The input text.
# output:
#   A string that is:
#   (1) the original text if the original text has all its letters written in uppercase (with the
#       allowed exception of the last letter being 's'),
#   (2) the lowercase version of the original text in other cases.
*/
func convertNonAbbreviationToLowercase(text string) string {
	// --------------------------------------------------------------------------------------------
	// Do not convert it if all its chars are capital (except that the last char is allowed to be a
	// lowercase 's')
	if len(text) > 1 {
		if text == strings.ToUpper(text) {
			return text
		}
		lenText := len(text)
		if text[lenText-1] == 's' && text[:lenText-1] == strings.ToUpper(text[:lenText-1]) {
			return text
		}
	}

	// --------------------------------------------------------------------------------------------
	// Otherwise convert it to lowercase.
	return strings.ToLower(text)
}

/*
# =================================================================================================
# function separateTextWithStopWords
# brief description:
#   Use stop words to seperate a sequence of words into candidate phrases.
# input:
#   phrases: a vector of word groups separated by puntuations.
# output:
#   a vector of candidate phrases
*/
func separateTextWithStopWords(phrases [][]string) [][]string {
	// --------------------------------------------------------------------------------------------
	// step 1: Prepare the result
	result := [][]string{make([]string, 0)}

	// --------------------------------------------------------------------------------------------
	// step 2: Convert the input word sequence into a sequence of candidate phrases
	for _, phrase := range phrases {
		// Must keep the phrases separated by puntuations
		if len(result[len(result)-1]) > 0 {
			result = append(result, make([]string, 0))
		}
		// Separate the phrases further using stop words
		for _, word := range phrase {
			_, isStopWord := stopWords[word]
			if isStopWord {
				if len(result[len(result)-1]) > 0 {
					result = append(result, make([]string, 0))
				}
			} else {
				result[len(result)-1] = append(result[len(result)-1], word)
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 3: Return the result
	if len(result[len(result)-1]) == 0 {
		newLen := len(result) - 1
		result = result[:newLen]
	}
	return result
}

/*
# =================================================================================================
# function separateHyphenedWords
# brief description:
#   Separate the hyphened words in candidate phrases into non-hyphened words for stemming later.
# input:
#   phrases: A vector of candidate phrases.
# output:
#   The unhyphenated candidate phrases.
*/
func separateHyphenedWords(phrases [][]string) [][]string {
	// --------------------------------------------------------------------------------------------
	// step 1: Prepare the result
	result := [][]string{}

	// --------------------------------------------------------------------------------------------
	// step 2: Separate the hyphened words in each candidate phrase
	for _, phrase := range phrases {
		unhyphenatedPhrase := []string{}
		for _, word := range phrase {
			if reHyphenedWords.MatchString(word) {
				subwords := strings.Split(word, "-")
				for _, subword := range subwords {
					unhyphenatedPhrase = append(unhyphenatedPhrase, subword)
				}
			} else {
				unhyphenatedPhrase = append(unhyphenatedPhrase, word)
			}
		}
		result = append(result, unhyphenatedPhrase)
	}

	// --------------------------------------------------------------------------------------------
	// step 3: Return the result
	return result
}

/*
# =================================================================================================
# function stemPhrases
# brief description:
#   Stem the words in each candidate phrases with Snowball stemmer (a.k.a. Porter 2 stemmer).
# input:
#   phrases: A vector of candidate phrases.
# output:
#   The stemmed candidate phrases.
# notes:
#   The reference to the stemmer used by us is:
#   Porter, M. F. (2001). Snowball: A language for stemming algorithms.
*/
func stemPhrases(phrases [][]string) []string {
	// --------------------------------------------------------------------------------------------
	// step 1: Prepare the result
	result := []string{}

	// --------------------------------------------------------------------------------------------
	// step 2: Stem the words in each candidate phrase
	for _, phrase := range phrases {
		stemmedPhrase := ""
		for _, word := range phrase {
			if len(stemmedPhrase) == 0 {
				stemmedPhrase = english.Stem(word, false)
			} else {
				stemmedPhrase += " " + english.Stem(word, false)
			}
		}
		result = append(result, stemmedPhrase)
	}

	// --------------------------------------------------------------------------------------------
	// step 3: Return the result
	return result
}

/*
# =================================================================================================
# function ExtractKeyPhraseCandidates
# brief description:
#   Search from the input text for key phrase candidates.
# input:
#   text: The input text.
# output:
#   A vector of the stems of the key phrase candidates.
*/
func ExtractKeyPhraseCandidates(text string) []string {
	// --------------------------------------------------------------------------------------------
	// step 1: Tokenize the input text into words.
	phrases := tokenizeIntoWords(text)

	// --------------------------------------------------------------------------------------------
	// step 2: Convert roman numbers to arabic numbers, then convert non-abbreviation words to lower
	//         case
	for idxPhrase, phrase := range phrases {
		for idxWord, word := range phrase {
			convertedWord := convertNonAbbreviationToLowercase(convertRomanToArabic(word))
			if convertedWord != word {
				phrases[idxPhrase][idxWord] = convertedWord
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 3: Use stop words to separate words into candidate phrases
	phrases = separateTextWithStopWords(phrases)

	// --------------------------------------------------------------------------------------------
	// step 4: Seperate the hyphened words in the phrases for stemming later
	phrases = separateHyphenedWords(phrases)

	// --------------------------------------------------------------------------------------------
	// step 5: Stem each phrase and return them
	result := stemPhrases(phrases)
	return result
}

/*
# =================================================================================================
# function TF
# brief description:
#	Compute Term Frequencies for a set of key phrase candidates with a set of auxiliary phrases
# input:
#	phraseCandidates: a set of key phrase candidates
#	auxPhrases: an array of auxiliary phrases
# output:
#	The term frequency
*/
func TF(phraseCandidates []string, auxPhrases []string) map[string]uint {
	// --------------------------------------------------------------------------------------------
	// step 1: initialize the result
	result := map[string]uint{}
	for _, candidate := range phraseCandidates {
		words := strings.Split(candidate, " ")
		numWords := len(words)
		for i := 0; i < numWords; i++ {
			text := words[i]
			result[text] = 0
			for j := i + 1; j < numWords; j++ {
				text += " " + words[j]
				result[text] = 0
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 2: scan through auxPhrases and compute the term frequencies
	for _, auxPhrase := range auxPhrases {
		auxWords := strings.Split(auxPhrase, " ")
		numAuxWords := len(auxWords)
		for i := 0; i < numAuxWords; i++ {
			text := auxWords[i]
			oldFreq, exists := result[text]
			if !exists {
				break
			}
			result[text] = oldFreq + 1

			for j := i + 1; j < numAuxWords; j++ {
				text += " " + auxWords[j]
				oldFreq, exists = result[text]
				if !exists {
					break
				}
				result[text] = oldFreq + 1
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step : return the result
	return result
}

/*
# =================================================================================================
# function IDF
# brief description:
#	Compute Inverse Document Frequencies from some sets of key phrase candidates
# input:
#	phraseCandidateGroups: some groups of key phrase candidates
# output:
#	the inverse document frequencies
*/
func IDF(phraseCandidateGroups [][]string) map[string]float64 {
	// --------------------------------------------------------------------------------------------
	// step 1: initialize the result
	result := map[string]float64{}

	// --------------------------------------------------------------------------------------------
	// step 2: count the document frequency
	for _, candidates := range phraseCandidateGroups {
		// first find the set of texts in this document
		groupResult := map[string]bool{}
		for _, candidate := range candidates {
			words := strings.Split(candidate, " ")
			numWords := len(words)
			for i := 0; i < numWords; i++ {
				text := words[i]
				groupResult[text] = true
				for j := i + 1; j < numWords; j++ {
					text += " " + words[j]
					groupResult[text] = true
				}
			}
		}

		// then update the document frequency with this set
		for text, _ := range groupResult {
			oldFreq, exists := result[text]
			if !exists {
				oldFreq = 0.0
			}
			result[text] = oldFreq + 1.0
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 3: compute inverse document frequency from document frequency
	n := len(phraseCandidateGroups)
	for text, df := range result {
		idf := math.Log(float64(n) / df)
		result[text] = idf
	}

	// --------------------------------------------------------------------------------------------
	// step 4: return the result
	return result
}

/*
# =================================================================================================
# function SimTF
# brief description:
#	Compute Fuzzy Term Frequencies for a set of key phrase candidates with a set of auxiliary phrases
# input:
#	phraseCandidates: a set of key phrase candidates
#	auxPhrases: an array of auxiliary phrases
#	phraseSimilarity: a sparse matrix that gives similarity between strings
# output:
#	The term frequency
*/
func SimTF(phraseCandidates []string, auxPhrases []string,
	phraseSimilarity map[string]map[string]float64) map[string]float64 {
	// --------------------------------------------------------------------------------------------
	// step 1: initialize the result
	result := map[string]float64{}
	for _, candidate := range phraseCandidates {
		words := strings.Split(candidate, " ")
		numWords := len(words)
		for i := 0; i < numWords; i++ {
			text := words[i]
			result[text] = 0
			for j := i + 1; j < numWords; j++ {
				text += " " + words[j]
				result[text] = 0.0
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 2: scan through auxPhrases and compute the term frequencies
	for _, auxPhrase := range auxPhrases {
		auxWords := strings.Split(auxPhrase, " ")
		numAuxWords := len(auxWords)
		for i := 0; i < numAuxWords; i++ {
			auxText := auxWords[i]
			auxSim, exists := phraseSimilarity[auxText]
			if !exists {
				continue
			}
			for text, oldFreq := range result {
				sim, exists := auxSim[text]
				if exists {
					result[text] = oldFreq + sim
				}
			}

			for j := i + 1; j < numAuxWords; j++ {
				auxText += " " + auxWords[j]
				auxSim, exists = phraseSimilarity[auxText]
				if !exists {
					break
				}
				for text, oldFreq := range result {
					sim, exists := auxSim[text]
					if exists {
						result[text] = oldFreq + sim
					}
				}
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step : return the result
	return result
}

/*
# =================================================================================================
# function SimIDF
# brief description:
#	Compute Fuzzy Inverse Document Frequencies from some sets of key phrase candidates
# input:
#	phraseCandidateGroups: some groups of key phrase candidates
#	phraseSimilarity: a sparse matrix that gives similarity between strings
# output:
#	the inverse document frequencies
*/
func SimIDF(phraseCandidateGroups [][]string, phraseSimilarity map[string]map[string]float64) map[string]float64 {
	// --------------------------------------------------------------------------------------------
	// step 1: initialize the result
	result := map[string]float64{}
	for _, candidates := range phraseCandidateGroups {
		for _, candidate := range candidates {
			words := strings.Split(candidate, " ")
			numWords := len(words)
			for i := 0; i < numWords; i++ {
				text := words[i]
				result[text] = 0.0
				for j := i + 1; j < numWords; j++ {
					text += " " + words[j]
					result[text] = 0.0
				}
			}
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 2: count the document frequency
	for idxGroup, candidates := range phraseCandidateGroups {
		// first initialize groupResult to those in candidates and those similar to the candidates
		groupResult := map[string]float64{}
		for _, candidate := range candidates {
			words := strings.Split(candidate, " ")
			numWords := len(words)
			for i := 0; i < numWords; i++ {
				text := words[i]
				groupResult[text] = 0.0
				simTexts, simExists := phraseSimilarity[text]
				if simExists {
					for simText, _ := range simTexts {
						_, resultExists := result[simText]
						if resultExists {
							groupResult[simText] = 0.0
						}
					}
				}
				for j := i + 1; j < numWords; j++ {
					text += " " + words[j]
					groupResult[text] = 0.0
					simTexts, simExists = phraseSimilarity[text]
					if simExists {
						for simText, _ := range simTexts {
							_, resultExists := result[simText]
							if resultExists {
								groupResult[simText] = 0.0
							}
						}
					}
				}
			}
		}

		// then find the set of texts in this document
		for _, candidate := range candidates {
			words := strings.Split(candidate, " ")
			numWords := len(words)

			for i := 0; i < numWords; i++ {
				text1 := words[i]
				for text2, oldValue := range groupResult {
					sim, exists := phraseSimilarity[text1][text2]
					if exists {
						groupResult[text2] = math.Max(oldValue, sim)
					}
				}
				for j := i + 1; j < numWords; j++ {
					text1 += " " + words[j]
					for text2, oldValue := range groupResult {
						sim, exists := phraseSimilarity[text1][text2]
						if exists {
							groupResult[text2] = math.Max(oldValue, sim)
						}
					}
				}
			}
		}

		// then update the document frequency with this set
		for text, value := range groupResult {
			result[text] += value
		}

		if (idxGroup+1)%1000 == 0 {
			fmt.Printf("%d of %d groups of sim IDF computed\n", idxGroup+1, len(phraseCandidateGroups))
		}
	}

	// --------------------------------------------------------------------------------------------
	// step 3: compute inverse document frequency from document frequency
	n := len(phraseCandidateGroups)
	for text, df := range result {
		idf := math.Log(float64(n) / df)
		result[text] = idf
	}

	// --------------------------------------------------------------------------------------------
	// step 4: return the result
	return result
}
