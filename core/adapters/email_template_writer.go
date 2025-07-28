package adapters

import (
	"bytes"
	"fmt"
	"text/template"
	"os"
	"encoding/json"
	"github.com/emersion/go-message/mail"
)
type EMLWriter struct {
	OutputDir	string
	Args		map[string]any
}
type EmailGroup struct {
	TO    []string `json:"TO"`
	CC    []string `json:"CC"`
	Dears []string `json:"Dears"`
}

type EmailDistribution map[string]EmailGroup

func (e *EMLWriter) BuildTmpl(tmpl string, raw map[string]any) (string, error) {
	strData := make(map[string]string, len(raw))
	for k, v := range raw {
		strData[k] = fmt.Sprintf("%v", v)
	}
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, strData); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *EMLWriter) BuildEmail(to, cc, dears []string) ([]byte, error) {
	
	header := mail.Header{}
	subject, ok := e.Args["subject"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'subject' in args")
	}
	header.SetSubject(subject)


	toList := make([]*mail.Address, len(to))
	for i, addr := range to {
		toList[i] = &mail.Address{Address: addr}
	}
	header.SetAddressList("To", toList)

	if len(cc) > 0 {
		ccList := make([]*mail.Address, len(cc))
		for i, addr := range cc {
			ccList[i] = &mail.Address{Address: addr}
		}
		header.SetAddressList("Cc", ccList)
	}
	return nil, nil
}

type BatchEmailTemplateWriter struct {
	AttachmentFilePaths  []string
	Template 			 string
	Writer				 *EMLWriter
	Emails	 			 EmailDistribution
}

func (e *BatchEmailTemplateWriter) Open() error {
	
	return nil
}	

func (e *BatchEmailTemplateWriter) LoadEmailDistributionFromFile(path string) (map[string]EmailGroup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var dist map[string]EmailGroup
	if err := json.Unmarshal(data, &dist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return dist, nil
}