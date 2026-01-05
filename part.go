package gimedic

// Part represents the Mozc/Gboard user dictionary POS tag.
// Values match mozc.user_dictionary.UserDictionary.PosType.
type Part int32

const (
	PartNone Part = iota
	PartNoun
	PartAbbreviation
	PartSuggestOnly
	PartProperNoun
	PartPersonName
	PartSurname
	PartGivenName
	PartOrganization
	PartPlaceName
	PartSuruNoun
	PartAdjectivalNoun
	PartNumber
	PartAlphabet
	PartSymbol
	PartEmoticon
	PartAdverb
	PartAdnominal
	PartConjunction
	PartInterjection
	PartPrefix
	PartCounter
	PartSuffixGeneral
	PartSuffixPersonName
	PartSuffixPlaceName
	PartVerbGodanWaRow
	PartVerbGodanKaRow
	PartVerbGodanSaRow
	PartVerbGodanTaRow
	PartVerbGodanNaRow
	PartVerbGodanMaRow
	PartVerbGodanRaRow
	PartVerbGodanGaRow
	PartVerbGodanBaRow
	PartVerbYodanHaRow
	PartVerbIchidan
	PartVerbKahen
	PartVerbSahen
	PartVerbZahen
	PartVerbRahen
	PartAdjective
	PartSentenceEndingParticle
	PartPunctuation
	PartFreeStandingWord
	PartSuppressionWord
)

var partNames = [...]string{
	"品詞なし",
	"名詞",
	"短縮よみ",
	"サジェストのみ",
	"固有名詞",
	"人名",
	"姓",
	"名",
	"組織",
	"地名",
	"名詞サ変",
	"名詞形動",
	"数",
	"アルファベット",
	"記号",
	"顔文字",
	"副詞",
	"連体詞",
	"接続詞",
	"感動詞",
	"接頭語",
	"助数詞",
	"接尾一般",
	"接尾人名",
	"接尾地名",
	"動詞ワ行五段",
	"動詞カ行五段",
	"動詞サ行五段",
	"動詞タ行五段",
	"動詞ナ行五段",
	"動詞マ行五段",
	"動詞ラ行五段",
	"動詞ガ行五段",
	"動詞バ行五段",
	"動詞ハ行四段",
	"動詞一段",
	"動詞カ変",
	"動詞サ変",
	"動詞ザ変",
	"動詞ラ変",
	"形容詞",
	"終助詞",
	"句読点",
	"独立語",
	"抑制単語",
}

func (p Part) String() string {
	if p < 0 || int(p) >= len(partNames) {
		return "unknown"
	}
	return partNames[p]
}

// ParsePart returns the Part matching the Japanese label.
func ParsePart(label string) (Part, bool) {
	for i, name := range partNames {
		if label == name {
			return Part(i), true
		}
	}
	return PartNone, false
}
