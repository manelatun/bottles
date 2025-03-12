package brew

import (
	"encoding/json"
	"os"
)

type Package struct {
	brew        *Brew
	Name        string
	FullName    string
	Version     string
	Bottles     set[string]
	PostInstall bool
}

func (p *Package) Bottle() {
	p.brew.Run("brew", "install", "--build-bottle", "--verbose", p.FullName)
	p.brew.Run("brew", "bottle", "--json", "--only-json-tab", "--verbose", p.FullName)

	bottleReceiptBlob, err := os.ReadFile(p.FullName + "--" + p.Version + ".catalina.bottle.json")
	if err != nil {
		p.brew.log.Fatal(err)
	}

	bottleReceipt := make(map[string]struct {
		Bottle struct {
			Tags struct {
				Catalina struct {
					LocalFilename string `json:"local_filename"`
				}
			}
		}
	})

	if err := json.Unmarshal(bottleReceiptBlob, &bottleReceipt); err != nil {
		p.brew.log.Fatal(err)
	}

	p.brew.Run("brew", "uninstall", "--verbose", p.FullName)
	p.brew.Run("brew", "install", "--verbose", bottleReceipt[p.FullName].Bottle.Tags.Catalina.LocalFilename)
	p.brew.Run("brew", "test", "--verbose", p.FullName)
}
