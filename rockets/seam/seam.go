package seam

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/esimov/caire"
	"github.com/pkg/errors"
)

func ContentAwareResize(url string) ([]byte, error) {
	fmt.Printf("Download starting for url %s\n", url)
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the image")
	}
	defer response.Body.Close()

	converted := &bytes.Buffer{}
	fmt.Printf("Download complete %s", url)

	shrinkfactor := 30
	fmt.Printf("Resize in progress %s, shrinking width by %d percent...\n", url, shrinkfactor)
	p := &caire.Processor{
		NewWidth:   shrinkfactor,
		Percentage: true,
	}

	err = p.Process(response.Body, converted)
	if err != nil {
		return nil, errors.Wrap(err, "failed to apply seam carving to the image")
	}
	fmt.Printf("Seam carving completed for %s\n", url)

	return converted.Bytes(), nil
}
