package registry

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type ABIErrorTranslator struct {
	abi         *abi.ABI
	translation map[string]*template.Template
}

// NewABIErrorTranslator constructs a new error translator from the given ABI and
// translation map. The keys of the map are names of errors in the ABI; the values
// are Go templates using variables named after the arguments of the corresponding
// error.
//
// For example, if our ABI had an error TooPoor(address addr), then our translation
// map could have the key "TooPoor" with value "Account {{ .addr }} is too poor.".
func NewABIErrorTranslator(abi *abi.ABI, translation map[string]string) (*ABIErrorTranslator, error) {
	m := make(map[string]*template.Template)

	for errName, message := range translation {
		_, ok := abi.Errors[errName]
		if !ok {
			return nil, fmt.Errorf("translation map references an error %s not in the ABI", errName)
		}

		t, err := template.New(errName).Parse(message)
		if err != nil {
			return nil, fmt.Errorf("error parsing translation template for error %s: %w", errName, err)
		}

		m[errName] = t
	}

	return &ABIErrorTranslator{abi: abi, translation: m}, nil
}

func (d *ABIErrorTranslator) Decode(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("length %d is too short, must have length at least 4", len(data))
	}

	selector := data[:4]
	argsData := data[4:]

	for _, abiErr := range d.abi.Errors {
		if bytes.Equal(selector, abiErr.ID[:4]) {
			errName := abiErr.Name

			message, ok := d.translation[abiErr.Name]
			if !ok {
				return "", fmt.Errorf("no translation for error %s", errName)
			}

			argMap := make(map[string]any, len(abiErr.Inputs))
			err := abiErr.Inputs.UnpackIntoMap(argMap, argsData)
			if err != nil {
				return "", fmt.Errorf("error unpacking arguments for error %s: %w", errName, err)
			}

			var b bytes.Buffer
			err = message.Execute(&b, argMap)
			if err != nil {
				return "", fmt.Errorf("error executing template for error %s: %w", errName, err)
			}

			return b.String(), nil
		}
	}

	return "", fmt.Errorf("unrecognized error with signature %s", hexutil.Encode(selector))
}
