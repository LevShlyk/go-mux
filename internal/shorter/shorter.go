package shorter

import "math"

const (
	DefaultAlphabetLength uint64 = 63
	DefaultShortLength    int    = 10
)

func getDefaultAlphabet() [DefaultAlphabetLength]string {
	return [DefaultAlphabetLength]string{
		"a", "b", "c", "d", "e", "f", "g", "h",
		"i", "j", "k", "l", "m", "n", "o", "p",
		"q", "r", "s", "t", "u", "v", "w", "x",
		"y", "z", "A", "B", "C", "D", "E", "F",
		"G", "H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U", "V",
		"W", "X", "Y", "Z", "1", "2", "3", "4",
		"5", "6", "7", "8", "9", "0", "_",
	}
}

type Shorter struct {
	alphabet       [DefaultAlphabetLength]string
	alphabetLength uint64
	shortLength    int
}

func BuildShorter() Shorter {
	return Shorter{
		alphabet:       getDefaultAlphabet(),
		alphabetLength: DefaultAlphabetLength,
		shortLength:    DefaultShortLength,
	}
}

func (s *Shorter) GetShortByID(id uint64) string {
	// первый символ алфавита
	char := uint64(0)
	// Минимальная длина строки - 10 символов
	index := id + uint64(math.Pow(float64(s.alphabetLength), float64(s.shortLength)))
	alpha := ""

	limit := 0
	for {
		part := char + (index % s.alphabetLength)
		alpha = s.alphabet[int(part)] + alpha
		index = (index / s.alphabetLength) >> 0
		limit++

		if index < 1 || limit >= s.shortLength {
			break
		}
	}

	return alpha
}
