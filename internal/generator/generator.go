package generator

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

type Batch = []ID

// Short is a 5 length array of runes. We could use strings
// here but easier to enforce length of code with this type.
type Short [5]rune

func (s Short) String() string {
	return string(s[:])
}

// ID is a 8 length array of bytes, which converted to hex
// would be a 16 character code.
type ID [8]byte

// NewIDFromHex will try to generate an ID from a given
// hex string.
func NewIDFromHex(h string) (ID, error) {
	bytes, err := hex.DecodeString(h)
	if err != nil {
		return ID{}, err
	}

	if len(bytes) != 8 {
		return ID{}, errors.New("hex value must be exactly 8 bytes")
	}

	id := ID{}
	copy(id[:], bytes[:])

	return id, nil
}

func (i ID) String() string {
	return strings.ToUpper(hex.EncodeToString(i[:]))
}

func (i ID) ShortForm() Short {

	// Encode the last 3 bytes of the ID representation to hex
	// then trim the first character, to get last 5 characters.
	// Then convert lower case hex representation to upper case.
	h := strings.ToUpper(hex.EncodeToString(i[len(i) - 3:])[1:])

	// Copy hex to rune array
	r := Short{}
	copy(r[:], []rune(h))

	return r
}

func (i ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

type RandomReader func([]byte) (int, error)

func (r RandomReader) Read(bytes []byte) (int, error) {
	return r(bytes)
}

type RandomGenerator interface {
	Read([]byte) (int, error)
}

type Generator struct {
	size int
	reader RandomGenerator
}

func NewGenerator(batchSize int, reader RandomGenerator) *Generator {
	return &Generator{
		size: batchSize,
		reader: reader,
	}
}

// Generate a slice of ID to given batch size.
// If anything went wrong generating ID, return an error.
func (g *Generator) Generate() ([]ID, error) {
	batch := make([]ID, g.size)

	// Map of Short to track short-form uniqueness.
	shorts := map[Short]struct{}{}

	for i, j := 0, len(batch); i < j; i++ {

		// Read random bytes into each ID
		_, err := g.reader.Read(batch[i][:])
		if err != nil {
			return nil, err
		}

		// Check if short-form is unique in batch,
		// if not then re-generate the ID.
		sf := batch[i].ShortForm()
		if _, ok := shorts[sf]; ok {
			i--
			continue
		}

		shorts[sf] = struct{}{}
	}

	return batch, nil
}