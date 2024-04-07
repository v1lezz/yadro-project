package words

import (
	"fmt"
	"testing"
)

func TestSnowBallStem_GetWordsFromTranscriptAndAlt(t *testing.T) {
	sbs := NewSnowBallStem()
	s := "[[ Person talks on phone.  Cat with many sharp points looks on. ]]\nPerson on phone: Hi, Dr. Elizabeth?  Yeah, uh ... I accidentally took the Fourier transform of my cat ...\nCat: Meow!\n{{alt-text: That cat has some serious periodic components}}"
	a := "That cat has some serious periodic components"
	fmt.Println(sbs.GetWordsFromTranscriptAndAlt(s, a))
}
