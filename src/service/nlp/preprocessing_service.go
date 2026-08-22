package nlp

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	reHTML       = regexp.MustCompile(`<[^>]+>`)
	reWhitespace = regexp.MustCompile(`\s+`)
	reURL        = regexp.MustCompile(`https?://\S+`)
	rePunct      = regexp.MustCompile(`[^\p{L}\p{N}\s]`)
)

// Indonesian stopwords — common function words that add no semantic value.
var stopwords = map[string]bool{
	"yang": true, "dan": true, "di": true, "ke": true, "dari": true,
	"ini": true, "itu": true, "dengan": true, "untuk": true, "pada": true,
	"dalam": true, "adalah": true, "akan": true, "juga": true, "tidak": true,
	"ada": true, "sudah": true, "atau": true, "oleh": true, "tersebut": true,
	"bisa": true, "lebih": true, "satu": true, "kita": true, "saya": true,
	"kami": true, "kamu": true, "dia": true, "mereka": true, "sebuah": true,
	"para": true, "hal": true, "bila": true, "jika": true, "maka": true,
	"serta": true, "tapi": true, "namun": true, "karena": true, "agar": true,
	"hingga": true, "seperti": true, "telah": true, "dapat": true, "bagi": true,
	"bahwa": true, "pun": true, "lagi": true, "jadi": true, "masih": true,
	"hanya": true, "sangat": true, "antara": true, "saat": true, "tahun": true,
	"no": true, "ia": true, "se": true,
}

type PreprocessingService struct{}

func NewPreprocessingService() *PreprocessingService {
	return &PreprocessingService{}
}

func (s *PreprocessingService) Clean(text string) string {
	text = reURL.ReplaceAllString(text, " ")
	text = reHTML.ReplaceAllString(text, " ")
	text = rePunct.ReplaceAllString(text, " ")
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, text)
	text = reWhitespace.ReplaceAllString(text, " ")
	return strings.TrimSpace(strings.ToLower(text))
}

func (s *PreprocessingService) RemoveStopwords(text string) string {
	words := strings.Fields(text)
	filtered := words[:0]
	for _, w := range words {
		if !stopwords[w] && len(w) > 1 {
			filtered = append(filtered, w)
		}
	}
	return strings.Join(filtered, " ")
}

func (s *PreprocessingService) Preprocess(rawText string) string {
	cleaned := s.Clean(rawText)
	return s.RemoveStopwords(cleaned)
}
